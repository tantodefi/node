package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	connectorzevm "github.com/zeta-chain/protocol-contracts/pkg/contracts/zevm/connectorzevm.sol"
	zrc20 "github.com/zeta-chain/protocol-contracts/pkg/contracts/zevm/zrc20.sol"
	"github.com/zeta-chain/node/cmd/zetacored/config"
	"github.com/zeta-chain/node/common"

	"github.com/zeta-chain/node/x/crosschain/types"
	fungibletypes "github.com/zeta-chain/node/x/fungible/types"
	zetaObserverTypes "github.com/zeta-chain/node/x/observer/types"
)

var _ evmtypes.EvmHooks = Hooks{}

type Hooks struct {
	k Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return h.k.PostTxProcessing(ctx, msg, receipt)
}

// PostTxProcessing implements EvmHooks.PostTxProcessing.
func (k Keeper) PostTxProcessing(
	ctx sdk.Context,
	msg core.Message,
	receipt *ethtypes.Receipt,
) error {
	abiStr := "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"}," +
		"{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"},{\"internalType\":\"uint8\",\"name\":\"decimals_\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"chainid_\",\"type\":\"uint256\"},{\"internalType\":\"enumCoinType\",\"name\":\"coinType_\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit_\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"systemContractAddress_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"CallerIsNotFungibleModule\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"GasFeeTransferFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSender\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LowAllowance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LowBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroGasCoin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroGasPrice\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"from\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"}],\"name\":\"UpdatedGasLimit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"protocolFlatFee\",\"type\":\"uint256\"}],\"name\":\"UpdatedProtocolFlatFee\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"systemContract\",\"type\":\"address\"}],\"name\":\"UpdatedSystemContract\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"to\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasfee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"protocolFlatFee\",\"type\":\"uint256\"}],\"name\":\"Withdrawal\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"CHAIN_ID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"COIN_TYPE\",\"outputs\":[{\"internalType\":\"enumCoinType\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"FUNGIBLE_MODULE_ADDRESS\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GAS_LIMIT\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PROTOCOL_FLAT_FEE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SYSTEM_CONTRACT_ADDRESS\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"}],\"name\":\"updateGasLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"protocolFlatFee\",\"type\":\"uint256\"}],\"name\":\"updateProtocolFlatFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"updateSystemContractAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"to\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdrawGasFee\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

	inputData := msg.Data()
	if len(inputData) >= 4 {
		// Check if method exist in ABI
		methodID := inputData[:4]
		parsedABI, err := abi.JSON(strings.NewReader(abiStr))
		if err != nil {
			return err
		}
		for _, method := range parsedABI.Methods {
			if bytes.Equal(methodID, method.ID) {
				// Check if deactivated method
				if method.Name == "increaseAllowance" || method.Name == "decreaseAllowance" {
					return fmt.Errorf("%s not allowed", method.Name)
				}
			}
		}
	}

	var emittingContract ethcommon.Address
	if msg.To() != nil {
		emittingContract = *msg.To()
	}
	return k.ProcessLogs(ctx, receipt.Logs, emittingContract, msg.From().Hex())
}

// ProcessLogs post-processes logs emitted by a zEVM contract; if the log contains Withdrawal event
// from registered ZRC20 contract, new CCTX will be created to trigger and track outbound
// transaction.
func (k Keeper) ProcessLogs(ctx sdk.Context, logs []*ethtypes.Log, emittingContract ethcommon.Address, txOrigin string) error {
	system, found := k.fungibleKeeper.GetSystemContract(ctx)
	if !found {
		return fmt.Errorf("cannot find system contract")
	}
	connectorZEVMAddr := ethcommon.HexToAddress(system.ConnectorZevm)
	if connectorZEVMAddr == (ethcommon.Address{}) {
		return fmt.Errorf("connectorZEVM address is empty")
	}

	for _, log := range logs {
		eventWithdrawal, err := k.ParseZRC20WithdrawalEvent(ctx, *log)
		if err == nil {
			if err := k.ProcessZRC20WithdrawalEvent(ctx, eventWithdrawal, emittingContract, txOrigin); err != nil {
				return err
			}
		}
		eZeta, err := ParseZetaSentEvent(*log, connectorZEVMAddr)
		if err == nil {
			if err := k.ProcessZetaSentEvent(ctx, eZeta, emittingContract, txOrigin); err != nil {
				return err
			}
		}
	}
	return nil
}

// ProcessZRC20WithdrawalEvent creates a new CCTX to process the withdrawal event
// error indicates system error and non-recoverable; should abort
func (k Keeper) ProcessZRC20WithdrawalEvent(ctx sdk.Context, event *zrc20.ZRC20Withdrawal, emittingContract ethcommon.Address, txOrigin string) error {
	if !k.zetaObserverKeeper.IsInboundEnabled(ctx) {
		return types.ErrNotEnoughPermissions
	}
	ctx.Logger().Info("ZRC20 withdrawal to %s amount %d\n", hex.EncodeToString(event.To), event.Value)
	tss, found := k.GetTSS(ctx)
	if !found {
		return errorsmod.Wrap(types.ErrCannotFindTSSKeys, "ProcessZRC20WithdrawalEvent: cannot be processed without TSS keys")
	}
	foreignCoin, found := k.fungibleKeeper.GetForeignCoins(ctx, event.Raw.Address.Hex())
	if !found {
		return fmt.Errorf("cannot find foreign coin with emittingContract address %s", event.Raw.Address.Hex())
	}

	receiverChain := k.zetaObserverKeeper.GetParams(ctx).GetChainFromChainID(foreignCoin.ForeignChainId)
	senderChain := common.ZetaChain()
	toAddr, err := receiverChain.EncodeAddress(event.To)
	if err != nil {
		return fmt.Errorf("cannot encode address %s: %s", event.To, err.Error())
	}

	gasLimit, err := k.fungibleKeeper.QueryGasLimit(ctx, ethcommon.HexToAddress(foreignCoin.Zrc20ContractAddress))
	if err != nil {
		return fmt.Errorf("cannot query gas limit: %s", err.Error())
	}

	// gasLimit+uint64(event.Raw.Index) to generate different cctx for multiple events in the same tx.
	msg := types.NewMsgVoteOnObservedInboundTx(
		"",
		emittingContract.Hex(),
		senderChain.ChainId,
		txOrigin,
		toAddr,
		foreignCoin.ForeignChainId,
		math.NewUintFromBigInt(event.Value),
		"",
		event.Raw.TxHash.String(),
		event.Raw.BlockNumber,
		gasLimit.Uint64()+uint64(event.Raw.Index),
		foreignCoin.CoinType,
		foreignCoin.Asset,
	)
	sendHash := msg.Digest()

	cctx := k.CreateNewCCTX(ctx, msg, sendHash, tss.TssPubkey, types.CctxStatus_PendingOutbound, &senderChain, receiverChain)

	// Get gas price and amount
	gasprice, found := k.GetGasPrice(ctx, receiverChain.ChainId)
	if !found {
		fmt.Printf("gasprice not found for %s\n", receiverChain)
		return fmt.Errorf("gasprice not found for %s", receiverChain)
	}
	cctx.GetCurrentOutTxParam().OutboundTxGasPrice = fmt.Sprintf("%d", gasprice.Prices[gasprice.MedianIndex])
	cctx.GetCurrentOutTxParam().Amount = cctx.InboundTxParams.Amount

	EmitZRCWithdrawCreated(ctx, cctx)
	return k.ProcessCCTX(ctx, cctx, receiverChain)
}

func (k Keeper) ProcessZetaSentEvent(ctx sdk.Context, event *connectorzevm.ZetaConnectorZEVMZetaSent, emittingContract ethcommon.Address, txOrigin string) error {
	if !k.zetaObserverKeeper.IsInboundEnabled(ctx) {
		return types.ErrNotEnoughPermissions
	}
	ctx.Logger().Info(fmt.Sprintf(
		"Zeta withdrawal to %s amount %d to chain with chainId %d",
		hex.EncodeToString(event.DestinationAddress),
		event.ZetaValueAndGas,
		event.DestinationChainId,
	))

	tss, found := k.GetTSS(ctx)
	if !found {
		return errorsmod.Wrap(types.ErrCannotFindTSSKeys, "ProcessZetaSentEvent: cannot be processed without TSS keys")
	}
	if err := k.bankKeeper.BurnCoins(
		ctx,
		fungibletypes.ModuleName,
		sdk.NewCoins(sdk.NewCoin(config.BaseDenom, sdk.NewIntFromBigInt(event.ZetaValueAndGas))),
	); err != nil {
		fmt.Printf("burn coins failed: %s\n", err.Error())
		return fmt.Errorf("ProcessZetaSentEvent: failed to burn coins from fungible: %s", err.Error())
	}

	receiverChainID := event.DestinationChainId
	receiverChain := k.zetaObserverKeeper.GetParams(ctx).GetChainFromChainID(receiverChainID.Int64())
	if receiverChain == nil {
		return zetaObserverTypes.ErrSupportedChains
	}
	// Validation if we want to send ZETA to an external chain, but there is no ZETA token.
	coreParams, found := k.zetaObserverKeeper.GetCoreParamsByChainID(ctx, receiverChain.ChainId)
	if !found {
		return types.ErrNotFoundCoreParams
	}
	if receiverChain.IsExternalChain() && coreParams.ZetaTokenContractAddress == "" {
		return types.ErrUnableToSendCoinType
	}
	toAddr := "0x" + hex.EncodeToString(event.DestinationAddress)
	senderChain := common.ZetaChain()
	amount := math.NewUintFromBigInt(event.ZetaValueAndGas)

	// Bump gasLimit by event index (which is very unlikely to be larger than 1000) to always have different ZetaSent events msgs.
	msg := types.NewMsgVoteOnObservedInboundTx(
		"",
		emittingContract.Hex(),
		senderChain.ChainId,
		txOrigin, toAddr,
		receiverChain.ChainId,
		amount,
		"",
		event.Raw.TxHash.String(),
		event.Raw.BlockNumber,
		90000+uint64(event.Raw.Index),
		common.CoinType_Zeta,
		"",
	)
	sendHash := msg.Digest()

	// Create the CCTX
	cctx := k.CreateNewCCTX(ctx, msg, sendHash, tss.TssPubkey, types.CctxStatus_PendingOutbound, &senderChain, receiverChain)

	if err := k.PayGasAndUpdateCctx(
		ctx,
		receiverChain.ChainId,
		&cctx,
		amount,
		true,
	); err != nil {
		return fmt.Errorf("ProcessWithdrawalEvent: pay gas failed: %s", err.Error())
	}

	EmitZetaWithdrawCreated(ctx, cctx)
	return k.ProcessCCTX(ctx, cctx, receiverChain)
}

func (k Keeper) ProcessCCTX(ctx sdk.Context, cctx types.CrossChainTx, receiverChain *common.Chain) error {
	inCctxIndex, ok := ctx.Value("inCctxIndex").(string)
	if ok {
		cctx.InboundTxParams.InboundTxObservedHash = inCctxIndex
	}

	if err := k.UpdateNonce(ctx, receiverChain.ChainId, &cctx); err != nil {
		return fmt.Errorf("ProcessWithdrawalEvent: update nonce failed: %s", err.Error())
	}

	k.SetCctxAndNonceToCctxAndInTxHashToCctx(ctx, cctx)
	ctx.Logger().Debug("ProcessCCTX successful \n")
	return nil
}

// ParseZRC20WithdrawalEvent tries extracting Withdrawal event from registered ZRC20 contract;
// returns error if the log entry is not a Withdrawal event, or is not emitted from a
// registered ZRC20 contract
func (k Keeper) ParseZRC20WithdrawalEvent(ctx sdk.Context, log ethtypes.Log) (*zrc20.ZRC20Withdrawal, error) {
	zrc20ZEVM, err := zrc20.NewZRC20Filterer(log.Address, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}
	event, err := zrc20ZEVM.ParseWithdrawal(log)
	if err != nil {
		return nil, err
	}

	coin, found := k.fungibleKeeper.GetForeignCoins(ctx, event.Raw.Address.Hex())
	if !found {
		return nil, fmt.Errorf("ParseZRC20WithdrawalEvent: cannot find foreign coin with contract address %s", event.Raw.Address.Hex())
	}
	chainID := coin.ForeignChainId
	if common.IsBitcoinChain(chainID) {
		if event.Value.Cmp(big.NewInt(0)) <= 0 {
			return nil, fmt.Errorf("ParseZRC20WithdrawalEvent: invalid amount %s", event.Value.String())
		}
		btcChainParams, err := common.GetBTCChainParams(chainID)
		if err != nil {
			return nil, err
		}
		addr, err := btcutil.DecodeAddress(string(event.To), btcChainParams)
		if err != nil {
			return nil, fmt.Errorf("ParseZRC20WithdrawalEvent: invalid address %s: %s", event.To, err)
		}
		_, ok := addr.(*btcutil.AddressWitnessPubKeyHash)
		if !ok {
			return nil, fmt.Errorf("ParseZRC20WithdrawalEvent: invalid address %s (not P2WPKH address)", event.To)
		}
	}
	return event, nil
}

func ParseZetaSentEvent(log ethtypes.Log, connectorZEVM ethcommon.Address) (*connectorzevm.ZetaConnectorZEVMZetaSent, error) {
	zetaConnectorZEVM, err := connectorzevm.NewZetaConnectorZEVMFilterer(log.Address, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}
	event, err := zetaConnectorZEVM.ParseZetaSent(log)
	if err != nil {
		return nil, err
	}

	if event.Raw.Address != connectorZEVM {
		return nil, fmt.Errorf("ParseZetaSentEvent: event address %s does not match connectorZEVM %s", event.Raw.Address.Hex(), connectorZEVM.Hex())
	}
	return event, nil
}
