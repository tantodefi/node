package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeta-chain/node/x/observer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SetKeygen set keygen in the store
func (k Keeper) SetKeygen(ctx sdk.Context, keygen types.Keygen) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeygenKey))
	b := k.cdc.MustMarshal(&keygen)
	store.Set([]byte{0}, b)
}

// GetKeygen returns keygen
func (k Keeper) GetKeygen(ctx sdk.Context) (val types.Keygen, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeygenKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveKeygen removes keygen from the store
func (k Keeper) RemoveKeygen(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeygenKey))
	store.Delete([]byte{0})
}

// Query

func (k Keeper) Keygen(c context.Context, _ *types.QueryGetKeygenRequest) (*types.QueryGetKeygenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	val, found := k.GetKeygen(ctx)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "not found")
	}
	return &types.QueryGetKeygenResponse{Keygen: &val}, nil
}
