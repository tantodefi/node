package keeper_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeta-chain/node/common"
	keepertest "github.com/zeta-chain/node/testutil/keeper"
	"github.com/zeta-chain/node/testutil/sample"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeta-chain/node/x/fungible/keeper"
	"github.com/zeta-chain/node/x/fungible/types"
)

func createNForeignCoins(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.ForeignCoins {
	items := make([]types.ForeignCoins, n)
	for i := range items {
		items[i].Zrc20ContractAddress = strconv.Itoa(i)

		keeper.SetForeignCoins(ctx, items[i])
	}
	return items
}

func setForeignCoins(ctx sdk.Context, k *keeper.Keeper, fc ...types.ForeignCoins) {
	for _, item := range fc {
		k.SetForeignCoins(ctx, item)
	}
}

func TestKeeper_GetGasCoinForForeignCoin(t *testing.T) {
	k, ctx, _, _ := keepertest.FungibleKeeper(t)

	// populate
	setForeignCoins(ctx, k,
		types.ForeignCoins{
			Zrc20ContractAddress: sample.EthAddress().String(),
			ForeignChainId:       1,
			CoinType:             common.CoinType_ERC20,
			Name:                 "foo",
		},
		types.ForeignCoins{
			Zrc20ContractAddress: sample.EthAddress().String(),
			ForeignChainId:       1,
			CoinType:             common.CoinType_ERC20,
			Name:                 "foo",
		},
		types.ForeignCoins{
			Zrc20ContractAddress: sample.EthAddress().String(),
			ForeignChainId:       1,
			CoinType:             common.CoinType_Gas,
			Name:                 "bar",
		},
		types.ForeignCoins{
			Zrc20ContractAddress: sample.EthAddress().String(),
			ForeignChainId:       2,
			CoinType:             common.CoinType_ERC20,
			Name:                 "foo",
		},
		types.ForeignCoins{
			Zrc20ContractAddress: sample.EthAddress().String(),
			ForeignChainId:       2,
			CoinType:             common.CoinType_ERC20,
			Name:                 "foo",
		},
	)

	fc, found := k.GetGasCoinForForeignCoin(ctx, 1)
	require.True(t, found)
	require.Equal(t, "bar", fc.Name)
	fc, found = k.GetGasCoinForForeignCoin(ctx, 2)
	require.False(t, found)
	fc, found = k.GetGasCoinForForeignCoin(ctx, 3)
	require.False(t, found)
}

func TestKeeperGetForeignCoinFromAsset(t *testing.T) {
	t.Run("can get foreign coin from asset", func(t *testing.T) {
		k, ctx, _, _ := keepertest.FungibleKeeper(t)

		gasAsset := sample.EthAddress().String()

		// populate
		setForeignCoins(ctx, k,
			types.ForeignCoins{
				Zrc20ContractAddress: sample.EthAddress().String(),
				Asset:                sample.EthAddress().String(),
				ForeignChainId:       1,
				CoinType:             common.CoinType_ERC20,
				Name:                 "foo",
			},
			types.ForeignCoins{
				Zrc20ContractAddress: sample.EthAddress().String(),
				Asset:                gasAsset,
				ForeignChainId:       1,
				CoinType:             common.CoinType_ERC20,
				Name:                 "bar",
			},
			types.ForeignCoins{
				Zrc20ContractAddress: sample.EthAddress().String(),
				Asset:                sample.EthAddress().String(),
				ForeignChainId:       1,
				CoinType:             common.CoinType_Gas,
				Name:                 "foo",
			},
			types.ForeignCoins{
				Zrc20ContractAddress: sample.EthAddress().String(),
				Asset:                sample.EthAddress().String(),
				ForeignChainId:       2,
				CoinType:             common.CoinType_ERC20,
				Name:                 "foo",
			},
			types.ForeignCoins{
				Zrc20ContractAddress: sample.EthAddress().String(),
				Asset:                sample.EthAddress().String(),
				ForeignChainId:       2,
				CoinType:             common.CoinType_ERC20,
				Name:                 "foo",
			},
		)

		fc, found := k.GetForeignCoinFromAsset(ctx, gasAsset, 1)
		require.True(t, found)
		require.Equal(t, "bar", fc.Name)
		fc, found = k.GetForeignCoinFromAsset(ctx, sample.EthAddress().String(), 1)
		require.False(t, found)
		fc, found = k.GetForeignCoinFromAsset(ctx, "invalid_address", 1)
		require.False(t, found)
		fc, found = k.GetForeignCoinFromAsset(ctx, gasAsset, 2)
		require.False(t, found)
		fc, found = k.GetForeignCoinFromAsset(ctx, gasAsset, 3)
		require.False(t, found)
	})

	t.Run("can get foreign coin with non-checksum address", func(t *testing.T) {
		k, ctx, _, _ := keepertest.FungibleKeeper(t)

		setForeignCoins(ctx, k,
			types.ForeignCoins{
				Zrc20ContractAddress: sample.EthAddress().String(),
				Asset:                "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
				ForeignChainId:       1,
				CoinType:             common.CoinType_ERC20,
				Name:                 "foo",
			},
		)

		fc, found := k.GetForeignCoinFromAsset(ctx, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", 1)
		require.True(t, found)
		require.Equal(t, "foo", fc.Name)
	})
}
