package zetaclient

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/rs/zerolog"
	"github.com/zeta-chain/node/common"
	"github.com/zeta-chain/node/x/crosschain/types"
	zetaObserverModuleTypes "github.com/zeta-chain/node/x/observer/types"
	"github.com/zeta-chain/node/zetaclient/config"
)

const (
	maxNoOfInputsPerTx = 20
	outTxBytesMin      = 400    // 500B is a conservative estimate for a 2-input, 3-output SegWit tx
	outTxBytesMax      = 4_000  // 4KB is a conservative estimate for a 21-input, 3-output SegWit tx
	outTxBytesCap      = 10_000 // in case of accident

	// for ZRC20 configuration
	bytesPerInput = 150                             // each input is about 150 bytes
	ZRC20GasLimit = outTxBytesMin + bytesPerInput*8 // 1600B a suggested ZRC20 GAS_LIMIT for a 10-input, 3-output SegWit tx
)

type BTCSigner struct {
	tssSigner TSSSigner
	rpcClient *rpcclient.Client
	logger    zerolog.Logger
	ts        *TelemetryServer
}

var _ ChainSigner = &BTCSigner{}

func NewBTCSigner(cfg config.BTCConfig, tssSigner TSSSigner, logger zerolog.Logger, ts *TelemetryServer) (*BTCSigner, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         cfg.RPCHost,
		User:         cfg.RPCUsername,
		Pass:         cfg.RPCPassword,
		HTTPPostMode: true,
		DisableTLS:   true,
		Params:       cfg.RPCParams,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating bitcoin rpc client: %s", err)
	}

	return &BTCSigner{
		tssSigner: tssSigner,
		rpcClient: client,
		logger: logger.With().
			Str("chain", "BTC").
			Str("module", "BTCSigner").Logger(),
		ts: ts,
	}, nil
}

// SignWithdrawTx receives utxos sorted by value, amount in BTC, feeRate in BTC per Kb
func (signer *BTCSigner) SignWithdrawTx(to *btcutil.AddressWitnessPubKeyHash, amount float64, gasPrice *big.Int, sizeLimit uint64,
	btcClient *BitcoinChainClient, height uint64, nonce uint64, chain *common.Chain) (*wire.MsgTx, error) {
	estimateFee := float64(gasPrice.Uint64()) * outTxBytesMax / 1e8
	nonceMark := common.NonceMarkAmount(nonce)

	// refresh unspent UTXOs and continue with keysign regardless of error
	err := btcClient.FetchUTXOS()
	if err != nil {
		signer.logger.Error().Err(err).Msgf("SignWithdrawTx: FetchUTXOS error: nonce %d chain %d", nonce, chain.ChainId)
	}

	// select N UTXOs to cover the total expense
	prevOuts, total, err := btcClient.SelectUTXOs(amount+estimateFee+float64(nonceMark)*1e-8, maxNoOfInputsPerTx, nonce, false)
	if err != nil {
		return nil, err
	}

	// build tx with selected unspents
	tx := wire.NewMsgTx(wire.TxVersion)
	for _, prevOut := range prevOuts {
		hash, err := chainhash.NewHashFromStr(prevOut.TxID)
		if err != nil {
			return nil, err
		}
		outpoint := wire.NewOutPoint(hash, prevOut.Vout)
		txIn := wire.NewTxIn(outpoint, nil, nil)
		tx.AddTxIn(txIn)
	}

	amountSatoshis, err := getSatoshis(amount)
	if err != nil {
		return nil, err
	}

	// size checking
	// #nosec G701 check as positive
	txSize := uint64(tx.SerializeSize())
	if txSize > sizeLimit { // ZRC20 'withdraw' charged less fee from end user
		signer.logger.Info().Msgf("sizeLimit %d is less than txSize %d for nonce %d", sizeLimit, txSize, nonce)
	}
	if txSize < outTxBytesMin { // outbound shouldn't be blocked a low sizeLimit
		signer.logger.Warn().Msgf("sizeLimit %d is less than outTxBytesMin %d; use outTxBytesMin", sizeLimit, outTxBytesMin)
		txSize = outTxBytesMin
	}
	if txSize > outTxBytesCap { // in case of accident
		signer.logger.Warn().Msgf("sizeLimit %d is greater than outTxBytesCap %d; use outTxBytesCap", sizeLimit, outTxBytesCap)
		txSize = outTxBytesCap
	}

	// fee calculation
	// #nosec G701 always in range (checked above)
	fees := new(big.Int).Mul(big.NewInt(int64(txSize)), gasPrice)
	fees.Div(fees, big.NewInt(bytesPerKB))
	signer.logger.Info().Msgf("bitcoin outTx nonce %d gasPrice %s size %d fees %s", nonce, gasPrice.String(), txSize, fees.String())

	// calculate remaining btc to TSS self
	tssAddrWPKH := signer.tssSigner.BTCAddressWitnessPubkeyHash()
	payToSelf, err := payToWitnessPubKeyHashScript(tssAddrWPKH.WitnessProgram())
	if err != nil {
		return nil, err
	}
	remaining := total - amount
	remainingSats, err := getSatoshis(remaining)
	if err != nil {
		return nil, err
	}
	remainingSats -= fees.Int64()
	remainingSats -= nonceMark
	if remainingSats < 0 {
		fmt.Printf("BTCSigner: SignWithdrawTx: Remainder Value is negative! : %d\n", remainingSats)
		fmt.Printf("BTCSigner: SignWithdrawTx: Number of inputs : %d\n", len(tx.TxIn))
		return nil, fmt.Errorf("remainder value is negative")
	} else if remainingSats == nonceMark {
		fmt.Printf("BTCSigner: SignWithdrawTx: Adjust remainder value to avoid duplicate nonce-mark: %d\n", remainingSats)
		remainingSats--
	}

	// 1st output: the nonce-mark btc to TSS self
	txOut1 := wire.NewTxOut(nonceMark, payToSelf)
	tx.AddTxOut(txOut1)

	// 2nd output: the payment to the recipient
	pkScript, err := payToWitnessPubKeyHashScript(to.WitnessProgram())
	if err != nil {
		return nil, err
	}
	txOut2 := wire.NewTxOut(amountSatoshis, pkScript)
	tx.AddTxOut(txOut2)

	// 3rd output: the remaining btc to TSS self
	if remainingSats > 0 {
		txOut3 := wire.NewTxOut(remainingSats, payToSelf)
		tx.AddTxOut(txOut3)
	}

	// sign the tx
	sigHashes := txscript.NewTxSigHashes(tx)
	witnessHashes := make([][]byte, len(tx.TxIn))
	for ix := range tx.TxIn {
		amt, err := getSatoshis(prevOuts[ix].Amount)
		if err != nil {
			return nil, err
		}
		pkScript, err := hex.DecodeString(prevOuts[ix].ScriptPubKey)
		if err != nil {
			return nil, err
		}
		witnessHashes[ix], err = txscript.CalcWitnessSigHash(pkScript, sigHashes, txscript.SigHashAll, tx, ix, amt)
		if err != nil {
			return nil, err
		}
	}
	tss, ok := signer.tssSigner.(*TSS)
	if !ok {
		return nil, fmt.Errorf("tssSigner is not a TSS")
	}
	sig65Bs, err := tss.SignBatch(witnessHashes, height, nonce, chain)
	if err != nil {
		return nil, fmt.Errorf("SignBatch error: %v", err)
	}

	for ix := range tx.TxIn {
		sig65B := sig65Bs[ix]
		R := big.NewInt(0).SetBytes(sig65B[:32])
		S := big.NewInt(0).SetBytes(sig65B[32:64])
		sig := btcec.Signature{
			R: R,
			S: S,
		}

		pkCompressed := signer.tssSigner.PubKeyCompressedBytes()
		hashType := txscript.SigHashAll
		txWitness := wire.TxWitness{append(sig.Serialize(), byte(hashType)), pkCompressed}
		tx.TxIn[ix].Witness = txWitness
	}
	return tx, nil
}

func (signer *BTCSigner) Broadcast(signedTx *wire.MsgTx) error {
	fmt.Printf("BTCSigner: Broadcasting: %s\n", signedTx.TxHash().String())

	var outBuff bytes.Buffer
	err := signedTx.Serialize(&outBuff)
	if err != nil {
		return err
	}
	str := hex.EncodeToString(outBuff.Bytes())
	fmt.Printf("BTCSigner: Transaction Data: %s\n", str)

	hash, err := signer.rpcClient.SendRawTransaction(signedTx, true)
	if err != nil {
		return err
	}
	signer.logger.Info().Msgf("Broadcasting BTC tx , hash %s ", hash)
	return nil
}

func (signer *BTCSigner) TryProcessOutTx(send *types.CrossChainTx, outTxMan *OutTxProcessorManager, outTxID string, chainclient ChainClient, zetaBridge *ZetaCoreBridge, height uint64) {
	defer func() {
		outTxMan.EndTryProcess(outTxID)
		if err := recover(); err != nil {
			signer.logger.Error().Msgf("BTC TryProcessOutTx: %s, caught panic error: %v", send.Index, err)
		}
	}()

	logger := signer.logger.With().
		Str("OutTxID", outTxID).
		Str("SendHash", send.Index).
		Logger()

	params := send.GetCurrentOutTxParam()
	if params.CoinType != common.CoinType_Gas {
		logger.Error().Msgf("BTC TryProcessOutTx: can only send BTC to a BTC network")
		return
	}

	logger.Info().Msgf("BTC TryProcessOutTx: %s, value %d to %s", send.Index, params.Amount.BigInt(), params.Receiver)
	btcClient, ok := chainclient.(*BitcoinChainClient)
	if !ok {
		logger.Error().Msgf("chain client is not a bitcoin client")
		return
	}
	flags, err := zetaBridge.GetCrosschainFlags()
	if err != nil {
		logger.Error().Err(err).Msgf("cannot get crosschain flags")
		return
	}
	if !flags.IsOutboundEnabled {
		logger.Info().Msgf("outbound is disabled")
		return
	}
	myid := zetaBridge.keys.GetAddress()
	// Early return if the send is already processed
	// FIXME: handle revert case
	outboundTxTssNonce := params.OutboundTxTssNonce
	included, confirmed, err := btcClient.IsSendOutTxProcessed(send.Index, outboundTxTssNonce, common.CoinType_Gas, logger)
	if err != nil {
		logger.Error().Err(err).Msgf("cannot check if send %s is processed", send.Index)
		return
	}
	if included || confirmed {
		logger.Info().Msgf("CCTX %s already processed; exit signer", outTxID)
		return
	}

	sizelimit := params.OutboundTxGasLimit
	gasprice, ok := new(big.Int).SetString(params.OutboundTxGasPrice, 10)
	if !ok || gasprice.Cmp(big.NewInt(0)) < 0 {
		logger.Error().Msgf("cannot convert gas price  %s ", params.OutboundTxGasPrice)
		return
	}

	// FIXME: config chain params
	addr, err := btcutil.DecodeAddress(params.Receiver, config.BitconNetParams)
	if err != nil {
		logger.Error().Err(err).Msgf("cannot decode address %s ", params.Receiver)
		return
	}
	to, ok := addr.(*btcutil.AddressWitnessPubKeyHash)
	if err != nil || !ok {
		logger.Error().Err(err).Msgf("cannot convert address %s to P2WPKH address", params.Receiver)
		return
	}

	logger.Info().Msgf("SignWithdrawTx: to %s, value %d sats", addr.EncodeAddress(), params.Amount.Uint64())
	logger.Info().Msgf("using utxos: %v", btcClient.utxos)
	tx, err := signer.SignWithdrawTx(to, float64(params.Amount.Uint64())/1e8, gasprice, sizelimit, btcClient, height,
		outboundTxTssNonce, &btcClient.chain)
	if err != nil {
		logger.Warn().Err(err).Msgf("SignOutboundTx error: nonce %d chain %d", outboundTxTssNonce, params.ReceiverChainId)
		return
	}
	logger.Info().Msgf("Key-sign success: %d => %s, nonce %d", send.InboundTxParams.SenderChainId, btcClient.chain.ChainName, outboundTxTssNonce)
	// FIXME: add prometheus metrics
	_, err = zetaBridge.GetObserverList(btcClient.chain)
	if err != nil {
		logger.Warn().Err(err).Msgf("unable to get observer list: chain %d observation %s", outboundTxTssNonce, zetaObserverModuleTypes.ObservationType_OutBoundTx.String())
	}
	if tx != nil {
		outTxHash := tx.TxHash().String()
		logger.Info().Msgf("on chain %s nonce %d, outTxHash %s signer %s", btcClient.chain.ChainName, outboundTxTssNonce, outTxHash, myid)
		// TODO: pick a few broadcasters.
		//if len(signers) == 0 || myid == signers[send.OutboundTxParams.Broadcaster] || myid == signers[int(send.OutboundTxParams.Broadcaster+1)%len(signers)] {
		// retry loop: 1s, 2s, 4s, 8s, 16s in case of RPC error
		for i := 0; i < 5; i++ {
			// #nosec G404 randomness is not a security issue here
			time.Sleep(time.Duration(rand.Intn(1500)) * time.Millisecond) //random delay to avoid sychronized broadcast
			err := signer.Broadcast(tx)
			if err != nil {
				logger.Warn().Err(err).Msgf("broadcasting tx %s to chain %s: nonce %d, retry %d", outTxHash, btcClient.chain.ChainName, outboundTxTssNonce, i)
				continue
			}
			logger.Info().Msgf("Broadcast success: nonce %d to chain %s outTxHash %s", outboundTxTssNonce, btcClient.chain.String(), outTxHash)
			zetaHash, err := zetaBridge.AddTxHashToOutTxTracker(btcClient.chain.ChainId, outboundTxTssNonce, outTxHash, nil, "", -1)
			if err != nil {
				logger.Err(err).Msgf("Unable to add to tracker on ZetaCore: nonce %d chain %s outTxHash %s", outboundTxTssNonce, btcClient.chain.ChainName, outTxHash)
			}
			logger.Info().Msgf("Broadcast to core successful %s", zetaHash)

			// Save successfully broadcasted transaction to btc chain client
			btcClient.SaveBroadcastedTx(outTxHash, outboundTxTssNonce)

			break // successful broadcast; no need to retry
		}
	}
}
