package keeper

import (
	"context"

	cosmoserror "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/zeta-chain/node/x/fungible/types"
	observertypes "github.com/zeta-chain/node/x/observer/types"
)

// UpdateContractBytecode updates the bytecode of a contract from the bytecode of an existing contract
// Only a ZRC20 contract or the WZeta connector contract can be updated
// IMPORTANT: the new contract bytecode must have the same storage layout as the old contract bytecode
// the new contract can add new variable but cannot remove any existing variable
func (k Keeper) UpdateContractBytecode(goCtx context.Context, msg *types.MsgUpdateContractBytecode) (*types.MsgUpdateContractBytecodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check authorization
	if msg.Creator != k.observerKeeper.GetParams(ctx).GetAdminPolicyAccount(observertypes.Policy_Type_group2) {
		return nil, cosmoserror.Wrap(sdkerrors.ErrUnauthorized, "Deploy can only be executed by the correct policy account")
	}

	// fetch account to update
	if !ethcommon.IsHexAddress(msg.ContractAddress) {
		return nil, cosmoserror.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", msg.ContractAddress)
	}
	contractAddress := ethcommon.HexToAddress(msg.ContractAddress)
	acct := k.evmKeeper.GetAccount(ctx, contractAddress)
	if acct == nil {
		return nil, cosmoserror.Wrapf(types.ErrContractNotFound, "contract (%s) not found", contractAddress.Hex())
	}

	// check the contract is a zrc20
	_, found := k.GetForeignCoins(ctx, msg.ContractAddress)
	if !found {
		// check contract is wzeta connector contract
		systemContract, found := k.GetSystemContract(ctx)
		if !found {
			return nil, types.ErrSystemContractNotFound
		}
		if msg.ContractAddress != systemContract.ConnectorZevm {
			// not a zrc20 or wzeta connector contract, can't be updated
			return nil, cosmoserror.Wrapf(types.ErrInvalidContract, "contract (%s) is neither a zrc20 nor wzeta connector", msg.ContractAddress)
		}
	}

	// fetch the account of the new bytecode
	if !ethcommon.IsHexAddress(msg.NewBytecodeAddress) {
		return nil, cosmoserror.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", msg.NewBytecodeAddress)
	}
	newBytecodeAddress := ethcommon.HexToAddress(msg.NewBytecodeAddress)
	newBytecodeAcct := k.evmKeeper.GetAccount(ctx, newBytecodeAddress)
	if newBytecodeAcct == nil {
		return nil, cosmoserror.Wrapf(types.ErrContractNotFound, "contract (%s) not found", newBytecodeAddress.Hex())
	}

	// set the new CodeHash to the account
	previousCodeHash := acct.CodeHash
	acct.CodeHash = newBytecodeAcct.CodeHash
	err := k.evmKeeper.SetAccount(ctx, contractAddress, *acct)
	if err != nil {
		return nil, cosmoserror.Wrapf(
			types.ErrSetBytecode,
			"failed to update contract (%s) bytecode (%s)",
			contractAddress.Hex(),
			err.Error(),
		)
	}
	k.Logger(ctx).Info(
		"updated contract bytecode",
		"contract", contractAddress.Hex(),
		"oldCodeHash", string(previousCodeHash),
		"newCodeHash", string(acct.CodeHash),
	)

	return &types.MsgUpdateContractBytecodeResponse{
		NewBytecodeHash: acct.CodeHash,
	}, nil
}
