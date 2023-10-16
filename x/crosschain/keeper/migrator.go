package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/zeta-chain/node/x/crosschain/migrations/v2"
	v3 "github.com/zeta-chain/node/x/crosschain/migrations/v3"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	crossChainKeeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{
		crossChainKeeper: keeper,
	}
}

// Migrate2to3 migrates the store from consensus version 2 to 3
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.crossChainKeeper.zetaObserverKeeper, m.crossChainKeeper.storeKey, m.crossChainKeeper.cdc)
}

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.crossChainKeeper.storeKey, m.crossChainKeeper.cdc)
}
