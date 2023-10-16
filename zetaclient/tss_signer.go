package zetaclient

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	observertypes "github.com/zeta-chain/node/x/observer/types"

	"github.com/zeta-chain/node/zetaclient/metrics"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	peer2 "github.com/libp2p/go-libp2p/core/peer"
	"github.com/zeta-chain/node/common"
	"github.com/zeta-chain/node/x/crosschain/types"
	"github.com/zeta-chain/node/zetaclient/config"
	"gitlab.com/thorchain/tss/go-tss/p2p"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/btcsuite/btcutil"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	zcommon "github.com/zeta-chain/node/common/cosmos"
	thorcommon "gitlab.com/thorchain/tss/go-tss/common"

	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/tss/go-tss/keysign"
	"gitlab.com/thorchain/tss/go-tss/tss"

	tmcrypto "github.com/tendermint/tendermint/crypto"
)

//var testPubKeys = []string{
//	"zetapub1addwnpepqtdklw8tf3anjz7nn5fly3uvq2e67w2apn560s4smmrt9e3x52nt2m5cmyy",
//	"zetapub1addwnpepqtspqyy6gk22u37ztra4hq3hdakc0w0k60sfy849mlml2vrpfr0wvszlzhs",
//	"zetapub1addwnpepq2ryyje5zr09lq7gqptjwnxqsy2vcdngvwd6z7yt5yjcnyj8c8cn5la9ezs",
//	"zetapub1addwnpepqfjcw5l4ay5t00c32mmlky7qrppepxzdlkcwfs2fd5u73qrwna0vzksjyd8",
//}
//
//var testPrivKeys = []string{
//	"MjQ1MDc2MmM4MjU5YjRhZjhhNmFjMmI0ZDBkNzBkOGE1ZTBmNDQ5NGI4NzM4OTYyM2E3MmI0OWMzNmE1ODZhNw==",
//	"YmNiMzA2ODU1NWNjMzk3NDE1OWMwMTM3MDU0NTNjN2YwMzYzZmVhZDE5NmU3NzRhOTMwOWIxN2QyZTQ0MzdkNg==",
//	"ZThiMDAxOTk2MDc4ODk3YWE0YThlMjdkMWY0NjA1MTAwZDgyNDkyYzdhNmMwZWQ3MDBhMWIyMjNmNGMzYjVhYg==",
//	"ZTc2ZjI5OTIwOGVlMDk2N2M3Yzc1MjYyODQ0OGUyMjE3NGJiOGRmNGQyZmVmODg0NzQwNmUzYTk1YmQyODlmNA==",
//}

type TSSKey struct {
	PubkeyInBytes  []byte // FIXME: compressed pubkey?
	PubkeyInBech32 string // FIXME: same above
	AddressInHex   string
}

func NewTSSKey(pk string) (*TSSKey, error) {
	TSSKey := &TSSKey{
		PubkeyInBech32: pk,
	}
	pubkey, err := zcommon.GetPubKeyFromBech32(zcommon.Bech32PubKeyTypeAccPub, pk)
	if err != nil {
		log.Error().Err(err).Msgf("GetPubKeyFromBech32 from %s", pk)
		return nil, fmt.Errorf("GetPubKeyFromBech32: %w", err)
	}
	decompresspubkey, err := crypto.DecompressPubkey(pubkey.Bytes())
	if err != nil {
		return nil, fmt.Errorf("NewTSS: DecompressPubkey error: %w", err)
	}
	TSSKey.PubkeyInBytes = crypto.FromECDSAPub(decompresspubkey)
	TSSKey.AddressInHex = crypto.PubkeyToAddress(*decompresspubkey).Hex()
	return TSSKey, nil
}

type TSS struct {
	Server        *tss.TssServer
	Keys          map[string]*TSSKey // PubkeyInBech32 => TSSKey
	CurrentPubkey string
	logger        zerolog.Logger
	Signers       []string
	CoreBridge    *ZetaCoreBridge
	Metrics       *ChainMetrics
}

var _ TSSSigner = (*TSS)(nil)

// FIXME: does it return pubkey in compressed form or uncompressed?
func (tss *TSS) Pubkey() []byte {
	return tss.Keys[tss.CurrentPubkey].PubkeyInBytes
}

// digest should be Hashes of some data
// Sign: Specify optionalPubkey to use a different pubkey than the current pubkey set during keygen
func (tss *TSS) Sign(digest []byte, height uint64, nonce uint64, chain *common.Chain, optionalPubKey string) ([65]byte, error) {
	H := digest
	log.Debug().Msgf("hash of digest is %s", H)

	tssPubkey := tss.CurrentPubkey
	if optionalPubKey != "" {
		tssPubkey = optionalPubKey
	}
	// #nosec G701 always in range
	keysignReq := keysign.NewRequest(tssPubkey, []string{base64.StdEncoding.EncodeToString(H)}, int64(height), nil, "0.14.0")
	ksRes, err := tss.Server.KeySign(keysignReq)
	if err != nil {
		log.Warn().Msg("keysign fail")
	}
	if ksRes.Status == thorcommon.Fail {
		log.Warn().Msgf("keysign status FAIL posting blame to core, blaming node(s): %#v", ksRes.Blame.BlameNodes)

		digest := hex.EncodeToString(digest)
		index := observertypes.GetBlameIndex(chain.ChainId, nonce, digest, height)

		zetaHash, err := tss.CoreBridge.PostBlameData(&ksRes.Blame, chain.ChainId, index)
		if err != nil {
			log.Error().Err(err).Msg("error sending blame data to core")
			return [65]byte{}, err
		}

		// Increment Blame counter
		for _, node := range ksRes.Blame.BlameNodes {
			counter, err := tss.Metrics.GetPromCounter(node.Pubkey)
			if err != nil {
				log.Error().Err(err).Msgf("error getting counter: %s", node.Pubkey)
				continue
			}
			counter.Inc()
		}

		log.Info().Msgf("keysign posted blame data tx hash: %s", zetaHash)
	}
	signature := ksRes.Signatures

	// [{cyP8i/UuCVfQKDsLr1kpg09/CeIHje1FU6GhfmyMD5Q= D4jXTH3/CSgCg+9kLjhhfnNo3ggy9DTQSlloe3bbKAs= eY++Z2LwsuKG1JcghChrsEJ4u9grLloaaFZNtXI3Ujk= AA==}]
	// 32B msg hash, 32B R, 32B S, 1B RC
	log.Info().Msgf("signature of digest is... %v", signature)

	if len(signature) == 0 {
		log.Warn().Err(err).Msgf("signature has length 0")
		return [65]byte{}, fmt.Errorf("keysign fail: %s", err)
	}
	if !verifySignature(tssPubkey, signature, H) {
		log.Error().Err(err).Msgf("signature verification failure")
		return [65]byte{}, fmt.Errorf("signuature verification fail")
	}
	var sigbyte [65]byte
	_, err = base64.StdEncoding.Decode(sigbyte[:32], []byte(signature[0].R))
	if err != nil {
		log.Error().Err(err).Msg("decoding signature R")
		return [65]byte{}, fmt.Errorf("signuature verification fail")
	}
	_, err = base64.StdEncoding.Decode(sigbyte[32:64], []byte(signature[0].S))
	if err != nil {
		log.Error().Err(err).Msg("decoding signature S")
		return [65]byte{}, fmt.Errorf("signuature verification fail")
	}
	_, err = base64.StdEncoding.Decode(sigbyte[64:65], []byte(signature[0].RecoveryID))
	if err != nil {
		log.Error().Err(err).Msg("decoding signature RecoveryID")
		return [65]byte{}, fmt.Errorf("signuature verification fail")
	}

	return sigbyte, nil
}

// digest should be batch of Hashes of some data
func (tss *TSS) SignBatch(digests [][]byte, height uint64, nonce uint64, chain *common.Chain) ([][65]byte, error) {
	tssPubkey := tss.CurrentPubkey
	digestBase64 := make([]string, len(digests))
	for i, digest := range digests {
		digestBase64[i] = base64.StdEncoding.EncodeToString(digest)
	}
	// #nosec G701 always in range
	keysignReq := keysign.NewRequest(tssPubkey, digestBase64, int64(height), nil, "0.14.0")

	ksRes, err := tss.Server.KeySign(keysignReq)
	if err != nil {
		log.Warn().Err(err).Msg("keysign fail")
	}

	if ksRes.Status == thorcommon.Fail {
		log.Warn().Msg("keysign status FAIL posting blame to core")
		digest := combineDigests(digestBase64)
		index := observertypes.GetBlameIndex(chain.ChainId, nonce, hex.EncodeToString(digest), height)

		zetaHash, err := tss.CoreBridge.PostBlameData(&ksRes.Blame, chain.ChainId, index)
		if err != nil {
			log.Error().Err(err).Msg("error sending blame data to core")
			return [][65]byte{}, err
		}

		// Increment Blame counter
		for _, node := range ksRes.Blame.BlameNodes {
			counter, err := tss.Metrics.GetPromCounter(node.Pubkey)
			if err != nil {
				log.Error().Err(err).Msgf("error getting counter: %s", node.Pubkey)
				continue
			}
			counter.Inc()
		}

		log.Info().Msgf("keysign posted blame data tx hash: %s", zetaHash)
	}

	signatures := ksRes.Signatures
	// [{cyP8i/UuCVfQKDsLr1kpg09/CeIHje1FU6GhfmyMD5Q= D4jXTH3/CSgCg+9kLjhhfnNo3ggy9DTQSlloe3bbKAs= eY++Z2LwsuKG1JcghChrsEJ4u9grLloaaFZNtXI3Ujk= AA==}]
	// 32B msg hash, 32B R, 32B S, 1B RC

	if len(signatures) != len(digests) {
		log.Warn().Err(err).Msgf("signature has length (%d) not equal to length of digests (%d)", len(signatures), len(digests))
		return [][65]byte{}, fmt.Errorf("keysign fail: %s", err)
	}

	//if !verifySignatures(tssPubkey, signatures, digests) {
	//	log.Error().Err(err).Msgf("signature verification failure")
	//	return [][65]byte{}, fmt.Errorf("signuature verification fail")
	//}
	pubkey, err := zcommon.GetPubKeyFromBech32(zcommon.Bech32PubKeyTypeAccPub, tssPubkey)
	if err != nil {
		log.Error().Msg("get pubkey from bech32 fail")
	}
	sigBytes := make([][65]byte, len(digests))
	for j, H := range digests {
		found := false
		D := base64.StdEncoding.EncodeToString(H)
		for _, signature := range signatures {
			if D == signature.Msg {
				found = true
				_, err = base64.StdEncoding.Decode(sigBytes[j][:32], []byte(signature.R))
				if err != nil {
					log.Error().Err(err).Msg("decoding signature R")
					return [][65]byte{}, fmt.Errorf("signuature verification fail")
				}
				_, err = base64.StdEncoding.Decode(sigBytes[j][32:64], []byte(signature.S))
				if err != nil {
					log.Error().Err(err).Msg("decoding signature S")
					return [][65]byte{}, fmt.Errorf("signuature verification fail")
				}
				_, err = base64.StdEncoding.Decode(sigBytes[j][64:65], []byte(signature.RecoveryID))
				if err != nil {
					log.Error().Err(err).Msg("decoding signature RecoveryID")
					return [][65]byte{}, fmt.Errorf("signuature verification fail")
				}
				sigPublicKey, err := crypto.SigToPub(H, sigBytes[j][:])
				if err != nil {
					log.Error().Err(err).Msg("SigToPub error in verify_signature")
					return [][65]byte{}, fmt.Errorf("signuature verification fail")
				}
				compressedPubkey := crypto.CompressPubkey(sigPublicKey)
				if bytes.Compare(pubkey.Bytes(), compressedPubkey) != 0 {
					log.Warn().Msgf("%d-th pubkey %s recovered pubkey %s", j, pubkey.String(), hex.EncodeToString(compressedPubkey))
					return [][65]byte{}, fmt.Errorf("signuature verification fail")
				}
			}
		}
		if !found {
			log.Error().Err(err).Msg("signature not found")
			return [][65]byte{}, fmt.Errorf("signuature verification fail")
		}
	}

	return sigBytes, nil
}

func (tss *TSS) Validate() error {
	evmAddress := tss.EVMAddress()
	blankAddress := ethcommon.Address{}
	if evmAddress == blankAddress {
		return fmt.Errorf("invalid evm address : %s", evmAddress.String())
	}
	if tss.BTCAddressWitnessPubkeyHash() == nil {
		return fmt.Errorf("invalid btc pub key hash : %s", tss.BTCAddress())
	}
	return nil
}

func (tss *TSS) EVMAddress() ethcommon.Address {
	addr, err := GetTssAddrEVM(tss.CurrentPubkey)
	if err != nil {
		log.Error().Err(err).Msg("getKeyAddr error")
		return ethcommon.Address{}
	}
	return addr
}

// generate a bech32 p2wpkh address from pubkey
func (tss *TSS) BTCAddress() string {
	addr, err := GetTssAddrBTC(tss.CurrentPubkey)
	if err != nil {
		log.Error().Err(err).Msg("getKeyAddr error")
		return ""
	}
	return addr
}

func (tss *TSS) BTCAddressWitnessPubkeyHash() *btcutil.AddressWitnessPubKeyHash {
	addrWPKH, err := getKeyAddrBTCWitnessPubkeyHash(tss.CurrentPubkey)
	if err != nil {
		log.Error().Err(err).Msg("BTCAddressPubkeyHash error")
		return nil
	}
	return addrWPKH
}

func (tss *TSS) PubKeyCompressedBytes() []byte {
	pubk, err := zcommon.GetPubKeyFromBech32(zcommon.Bech32PubKeyTypeAccPub, tss.CurrentPubkey)
	if err != nil {
		log.Error().Err(err).Msg("PubKeyCompressedBytes error")
		return nil
	}
	return pubk.Bytes()
}

// adds a new key to the TSS keys map
func (tss *TSS) InsertPubKey(pk string) error {
	TSSKey, err := NewTSSKey(pk)
	if err != nil {
		return err
	}
	tss.Keys[pk] = TSSKey
	return nil
}

func GetTssAddrEVM(tssPubkey string) (ethcommon.Address, error) {
	var keyAddr ethcommon.Address
	pubk, err := zcommon.GetPubKeyFromBech32(zcommon.Bech32PubKeyTypeAccPub, tssPubkey)
	if err != nil {
		log.Fatal().Err(err)
		return keyAddr, err
	}
	//keyAddrBytes := pubk.EVMAddress().Bytes()
	pubk.Bytes()
	decompresspubkey, err := crypto.DecompressPubkey(pubk.Bytes())
	if err != nil {
		log.Fatal().Err(err).Msg("decompress err")
		return keyAddr, err
	}

	keyAddr = crypto.PubkeyToAddress(*decompresspubkey)

	return keyAddr, nil
}

// FIXME: mainnet/testnet
func GetTssAddrBTC(tssPubkey string) (string, error) {
	addrWPKH, err := getKeyAddrBTCWitnessPubkeyHash(tssPubkey)
	if err != nil {
		log.Fatal().Err(err)
		return "", err
	}

	return addrWPKH.EncodeAddress(), nil
}

func getKeyAddrBTCWitnessPubkeyHash(tssPubkey string) (*btcutil.AddressWitnessPubKeyHash, error) {
	pubk, err := zcommon.GetPubKeyFromBech32(zcommon.Bech32PubKeyTypeAccPub, tssPubkey)
	if err != nil {
		return nil, err
	}
	addr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pubk.Bytes()), config.BitconNetParams)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func NewTSS(peer p2p.AddrList, privkey tmcrypto.PrivKey, preParams *keygen.LocalPreParams, cfg *config.Config, bridge *ZetaCoreBridge, tssHistoricalList []types.TSS, metrics *metrics.Metrics) (*TSS, error) {
	server, err := SetupTSSServer(peer, privkey, preParams, cfg)
	if err != nil {
		return nil, fmt.Errorf("SetupTSSServer error: %w", err)
	}
	newTss := TSS{
		Server:        server,
		Keys:          make(map[string]*TSSKey),
		CurrentPubkey: cfg.CurrentTssPubkey,
		logger:        log.With().Str("module", "tss_signer").Logger(),
		CoreBridge:    bridge,
	}

	err = newTss.LoadTssFilesFromDirectory(cfg.TssPath)
	if err != nil {
		return nil, err
	}
	_, pubkeyInBech32, err := GetKeyringKeybase(cfg)
	if err != nil {
		return nil, err
	}
	err = newTss.VerifyKeysharesForPubkeys(tssHistoricalList, pubkeyInBech32)
	if err != nil {
		bridge.logger.Error().Err(err).Msg("VerifyKeysharesForPubkeys fail")
	}
	err = newTss.RegisterMetrics(metrics)
	if err != nil {
		bridge.logger.Err(err).Msg("tss.RegisterMetrics")
		return nil, err
	}

	return &newTss, nil
}

func (tss *TSS) VerifyKeysharesForPubkeys(tssList []types.TSS, granteePubKey32 string) error {
	for _, t := range tssList {
		if WasNodePartOfTss(granteePubKey32, t.TssParticipantList) {
			if _, ok := tss.Keys[t.TssPubkey]; !ok {
				return fmt.Errorf("pubkey %s not found in keyshare", t.TssPubkey)
			}
		}
	}
	return nil
}

func WasNodePartOfTss(granteePubKey32 string, granteeList []string) bool {
	for _, grantee := range granteeList {
		if granteePubKey32 == grantee {
			return true
		}
	}
	return false
}
func (tss *TSS) LoadTssFilesFromDirectory(tssPath string) error {
	files, err := os.ReadDir(tssPath)
	if err != nil {
		fmt.Println("ReadDir error :", err.Error())
		return err
	}
	found := false
	var sharefiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(filepath.Base(file.Name()), "localstate") {
			sharefiles = append(sharefiles, file)
		}
	}
	if len(sharefiles) > 0 {
		sort.SliceStable(sharefiles, func(i, j int) bool {
			fi, err := sharefiles[i].Info()
			if err != nil {
				return false
			}
			fj, err := sharefiles[j].Info()
			if err != nil {
				return false
			}
			return fi.ModTime().After(fj.ModTime())
		})
		tss.logger.Info().Msgf("found %d localstate files", len(sharefiles))
		for _, localStateFile := range sharefiles {
			filename := filepath.Base(localStateFile.Name())
			filearray := strings.Split(filename, "-")
			if len(filearray) == 2 {
				log.Info().Msgf("Found stored Pubkey in local state: %s", filearray[1])
				pk := strings.TrimSuffix(filearray[1], ".json")

				err = tss.InsertPubKey(pk)
				if err != nil {
					log.Error().Err(err).Msg("InsertPubKey  in NewTSS fail")
				}
				tss.logger.Info().Msgf("registering TSS pubkey %s (eth hex %s)", pk, tss.Keys[pk].AddressInHex)
				found = true
			}
		}
	}
	if !found {
		log.Info().Msg("TSS Keyshare file NOT found")
	}
	return nil
}

func SetupTSSServer(peer p2p.AddrList, privkey tmcrypto.PrivKey, preParams *keygen.LocalPreParams, cfg *config.Config) (*tss.TssServer, error) {
	bootstrapPeers := peer
	log.Info().Msgf("Peers AddrList %v", bootstrapPeers)

	tsspath := cfg.TssPath
	if len(tsspath) == 0 {
		log.Error().Msg("empty env TSSPATH")
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Error().Err(err).Msgf("cannot get UserHomeDir")
			return nil, err
		}
		tsspath = path.Join(homedir, ".Tss")
		log.Info().Msgf("create temporary TSSPATH: %s", tsspath)
	}
	IP := cfg.PublicIP
	if len(IP) == 0 {
		log.Info().Msg("empty public IP in config")
	}
	tssServer, err := tss.NewTss(
		bootstrapPeers,
		6668,
		privkey,
		"MetaMetaOpenTheDoor",
		tsspath,
		thorcommon.TssConfig{
			EnableMonitor:   true,
			KeyGenTimeout:   300 * time.Second, // must be shorter than constants.JailTimeKeygen
			KeySignTimeout:  30 * time.Second,  // must be shorter than constants.JailTimeKeysign
			PartyTimeout:    30 * time.Second,
			PreParamTimeout: 5 * time.Minute,
		},
		preParams, // use pre-generated pre-params if non-nil
		IP,        // for docker test
	)
	if err != nil {
		log.Error().Err(err).Msg("NewTSS error")
		return nil, fmt.Errorf("NewTSS error: %w", err)
	}

	err = tssServer.Start()
	if err != nil {
		log.Error().Err(err).Msg("tss server start error")
	}

	log.Info().Msgf("LocalID: %v", tssServer.GetLocalPeerID())
	if tssServer.GetLocalPeerID() == "" ||
		tssServer.GetLocalPeerID() == "0" ||
		tssServer.GetLocalPeerID() == "000000000000000000000000000000" ||
		tssServer.GetLocalPeerID() == peer2.ID("").String() {
		log.Error().Msg("tss server start error")
		return nil, fmt.Errorf("tss server start error")
	}

	return tssServer, nil
}

func TestKeysign(tssPubkey string, tssServer *tss.TssServer) error {
	log.Info().Msg("trying keysign...")
	data := []byte("hello meta")
	H := crypto.Keccak256Hash(data)
	log.Info().Msgf("hash of data (hello meta) is %s", H)

	keysignReq := keysign.NewRequest(tssPubkey, []string{base64.StdEncoding.EncodeToString(H.Bytes())}, 10, nil, "0.14.0")
	ksRes, err := tssServer.KeySign(keysignReq)
	if err != nil {
		log.Warn().Msg("keysign fail")
	}
	signature := ksRes.Signatures
	// [{cyP8i/UuCVfQKDsLr1kpg09/CeIHje1FU6GhfmyMD5Q= D4jXTH3/CSgCg+9kLjhhfnNo3ggy9DTQSlloe3bbKAs= eY++Z2LwsuKG1JcghChrsEJ4u9grLloaaFZNtXI3Ujk= AA==}]
	// 32B msg hash, 32B R, 32B S, 1B RC
	log.Info().Msgf("signature of helloworld... %v", signature)

	if len(signature) == 0 {
		log.Info().Msgf("signature has length 0, skipping verify")
		return fmt.Errorf("signature has length 0")
	}
	verifySignature(tssPubkey, signature, H.Bytes())
	if verifySignature(tssPubkey, signature, H.Bytes()) {
		return nil
	}
	return fmt.Errorf("verify signature fail")
}

func verifySignature(tssPubkey string, signature []keysign.Signature, H []byte) bool {
	if len(signature) == 0 {
		log.Warn().Msg("verify_signature: empty signature array")
		return false
	}
	pubkey, err := zcommon.GetPubKeyFromBech32(zcommon.Bech32PubKeyTypeAccPub, tssPubkey)
	if err != nil {
		log.Error().Msg("get pubkey from bech32 fail")
	}
	// verify the signature of msg.
	var sigbyte [65]byte
	_, err = base64.StdEncoding.Decode(sigbyte[:32], []byte(signature[0].R))
	if err != nil {
		log.Error().Err(err).Msg("decoding signature R")
		return false
	}
	_, err = base64.StdEncoding.Decode(sigbyte[32:64], []byte(signature[0].S))
	if err != nil {
		log.Error().Err(err).Msg("decoding signature S")
		return false
	}
	_, err = base64.StdEncoding.Decode(sigbyte[64:65], []byte(signature[0].RecoveryID))
	if err != nil {
		log.Error().Err(err).Msg("decoding signature RecoveryID")
		return false
	}
	sigPublicKey, err := crypto.SigToPub(H, sigbyte[:])
	if err != nil {
		log.Error().Err(err).Msg("SigToPub error in verify_signature")
		return false
	}
	compressedPubkey := crypto.CompressPubkey(sigPublicKey)
	log.Info().Msgf("pubkey %s recovered pubkey %s", pubkey.String(), hex.EncodeToString(compressedPubkey))
	return bytes.Compare(pubkey.Bytes(), compressedPubkey) == 0
}

func combineDigests(digestList []string) []byte {
	digestConcat := strings.Join(digestList[:], "")
	digestBytes := chainhash.DoubleHashH([]byte(digestConcat))
	return digestBytes.CloneBytes()
}

func (tss *TSS) RegisterMetrics(metrics *metrics.Metrics) error {
	tss.Metrics = NewChainMetrics("tss", metrics)
	keygenRes, err := tss.CoreBridge.GetKeyGen()
	if err != nil {
		return err
	}
	for _, key := range keygenRes.GranteePubkeys {
		err := tss.Metrics.RegisterPromCounter(key, "tss node blame counter")
		if err != nil {
			return err
		}
	}
	return nil
}
