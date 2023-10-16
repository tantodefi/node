package emissions_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "github.com/zeta-chain/node/testutil/keeper"
	"github.com/zeta-chain/node/testutil/nullify"
	"github.com/zeta-chain/node/testutil/sample"
	"github.com/zeta-chain/node/x/emissions"
	"github.com/zeta-chain/node/x/emissions/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		WithdrawableEmissions: []types.WithdrawableEmissions{
			sample.WithdrawableEmissions(t),
			sample.WithdrawableEmissions(t),
			sample.WithdrawableEmissions(t),
		},
	}

	// Init and export
	k, ctx := keepertest.EmissionsKeeper(t)
	emissions.InitGenesis(ctx, *k, genesisState)
	got := emissions.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	// Compare genesis after init and export
	nullify.Fill(&genesisState)
	nullify.Fill(got)
	require.Equal(t, genesisState, *got)
}
