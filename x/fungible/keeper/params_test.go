package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/zeta-chain/node/testutil/keeper"
	"github.com/zeta-chain/node/x/fungible/types"
)

func TestGetParams(t *testing.T) {
	k, ctx, _, _ := testkeeper.FungibleKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
