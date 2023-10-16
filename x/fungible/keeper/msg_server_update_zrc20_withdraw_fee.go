package keeper

import (
	"context"

	cosmoserrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/zeta-chain/node/x/fungible/types"
	zetaObserverTypes "github.com/zeta-chain/node/x/observer/types"
)

func (k Keeper) UpdateZRC20WithdrawFee(goCtx context.Context, msg *types.MsgUpdateZRC20WithdrawFee) (*types.MsgUpdateZRC20WithdrawFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check signer permission
	if msg.Creator != k.observerKeeper.GetParams(ctx).GetAdminPolicyAccount(zetaObserverTypes.Policy_Type_group2) {
		return nil, cosmoserrors.Wrap(sdkerrors.ErrUnauthorized, "deploy can only be executed by the correct policy account")
	}

	// check the zrc20 exists
	zrc20Addr := ethcommon.HexToAddress(msg.Zrc20Address)
	if zrc20Addr == (ethcommon.Address{}) {
		return nil, cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid zrc20 contract address (%s)", msg.Zrc20Address)
	}
	coin, found := k.GetForeignCoins(ctx, msg.Zrc20Address)
	if !found {
		return nil, cosmoserrors.Wrapf(types.ErrForeignCoinNotFound, "no foreign coin match requested zrc20 address (%s)", msg.Zrc20Address)
	}

	// get the previous fee
	oldWithdrawFee, err := k.QueryProtocolFlatFee(ctx, zrc20Addr)
	if err != nil {
		return nil, cosmoserrors.Wrapf(types.ErrContractCall, "failed to query protocol flat fee (%s)", err.Error())
	}
	oldGasLimit, err := k.QueryGasLimit(ctx, zrc20Addr)
	if err != nil {
		return nil, cosmoserrors.Wrapf(types.ErrContractCall, "failed to query gas limit (%s)", err.Error())
	}

	// call the contract methods
	tmpCtx, commit := ctx.CacheContext()
	if !msg.NewWithdrawFee.IsNil() {
		_, err = k.UpdateZRC20ProtocolFlatFee(tmpCtx, zrc20Addr, msg.NewWithdrawFee.BigInt())
		if err != nil {
			return nil, cosmoserrors.Wrapf(types.ErrContractCall, "failed to call zrc20 contract updateProtocolFlatFee method (%s)", err.Error())
		}
	}
	if !msg.NewGasLimit.IsNil() {
		_, err = k.UpdateZRC20GasLimit(tmpCtx, zrc20Addr, msg.NewGasLimit.BigInt())
		if err != nil {
			return nil, cosmoserrors.Wrapf(types.ErrContractCall, "failed to call zrc20 contract updateGasLimit method (%s)", err.Error())
		}
	}

	err = ctx.EventManager().EmitTypedEvent(
		&types.EventZRC20WithdrawFeeUpdated{
			MsgTypeUrl:     sdk.MsgTypeURL(&types.MsgUpdateZRC20WithdrawFee{}),
			ChainId:        coin.ForeignChainId,
			CoinType:       coin.CoinType,
			Zrc20Address:   zrc20Addr.Hex(),
			OldWithdrawFee: oldWithdrawFee.String(),
			NewWithdrawFee: msg.NewWithdrawFee.String(),
			Signer:         msg.Creator,
			OldGasLimit:    oldGasLimit.String(),
			NewGasLimit:    msg.NewGasLimit.String(),
		},
	)
	if err != nil {
		k.Logger(ctx).Error("failed to emit event", "error", err.Error())
		return nil, cosmoserrors.Wrapf(types.ErrEmitEvent, "failed to emit event (%s)", err.Error())
	}
	commit()

	return &types.MsgUpdateZRC20WithdrawFeeResponse{}, nil
}
