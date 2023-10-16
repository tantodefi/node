package v3_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "github.com/zeta-chain/node/testutil/keeper"
	"github.com/zeta-chain/node/testutil/sample"
	v3 "github.com/zeta-chain/node/x/observer/migrations/v3"
	"github.com/zeta-chain/node/x/observer/types"
)

func TestMigrateStore(t *testing.T) {
	k, ctx := keepertest.ObserverKeeper(t)

	// nothing if no admin policy
	params := types.DefaultParams()
	params.AdminPolicy = []*types.Admin_Policy{}
	k.SetParams(ctx, params)
	err := v3.MigrateStore(ctx, k)
	require.NoError(t, err)
	params = k.GetParams(ctx)
	require.Len(t, params.AdminPolicy, 0)

	// update admin policy
	admin := sample.AccAddress()
	params = types.DefaultParams()
	params.AdminPolicy = []*types.Admin_Policy{
		{
			Address:    admin,
			PolicyType: 0,
		},
		{
			Address:    sample.AccAddress(),
			PolicyType: 5,
		},
		{
			Address:    admin,
			PolicyType: 10,
		},
	}
	k.SetParams(ctx, params)
	err = v3.MigrateStore(ctx, k)
	require.NoError(t, err)
	params = k.GetParams(ctx)
	require.Len(t, params.AdminPolicy, 2)
	require.Equal(t, params.AdminPolicy[0].PolicyType, types.Policy_Type_group1)
	require.Equal(t, params.AdminPolicy[1].PolicyType, types.Policy_Type_group2)
	require.Equal(t, params.AdminPolicy[0].Address, admin)
	require.Equal(t, params.AdminPolicy[1].Address, admin)
}
