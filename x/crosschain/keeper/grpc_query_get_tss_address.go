package keeper

import (
	"context"

	"github.com/btcsuite/btcutil"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	zcommon "github.com/zeta-chain/node/common/cosmos"
	"github.com/zeta-chain/node/zetaclient/config"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeta-chain/node/x/crosschain/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetTssAddress(goCtx context.Context, req *types.QueryGetTssAddressRequest) (*types.QueryGetTssAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	tss, found := k.GetTSS(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}
	ethAddress, err := getTssAddrEVM(tss.TssPubkey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	btcAddress, err := getTssAddrBTC(tss.TssPubkey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetTssAddressResponse{
		Eth: ethAddress.String(),
		Btc: btcAddress,
	}, nil
}

func getTssAddrEVM(tssPubkey string) (ethcommon.Address, error) {
	var keyAddr ethcommon.Address
	pubk, err := zcommon.GetPubKeyFromBech32(zcommon.Bech32PubKeyTypeAccPub, tssPubkey)
	if err != nil {
		return keyAddr, err
	}
	//keyAddrBytes := pubk.EVMAddress().Bytes()
	pubk.Bytes()
	decompresspubkey, err := crypto.DecompressPubkey(pubk.Bytes())
	if err != nil {
		return keyAddr, err
	}

	keyAddr = crypto.PubkeyToAddress(*decompresspubkey)

	return keyAddr, nil
}

func getTssAddrBTC(tssPubkey string) (string, error) {
	addrWPKH, err := getKeyAddrBTCWitnessPubkeyHash(tssPubkey)
	if err != nil {
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
