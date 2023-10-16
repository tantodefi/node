package emissions

import (
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeta-chain/node/cmd/zetacored/config"
	"github.com/zeta-chain/node/x/emissions/keeper"
	"github.com/zeta-chain/node/x/emissions/types"
)

func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	reservesFactor, bondFactor, durationFactor := keeper.GetBlockRewardComponents(ctx)
	blockRewards := reservesFactor.Mul(bondFactor).Mul(durationFactor)
	if blockRewards.IsZero() {
		return
	}
	validatorRewards := sdk.MustNewDecFromStr(keeper.GetParams(ctx).ValidatorEmissionPercentage).Mul(blockRewards).TruncateInt()
	observerRewards := sdk.MustNewDecFromStr(keeper.GetParams(ctx).ObserverEmissionPercentage).Mul(blockRewards).TruncateInt()
	tssSignerRewards := sdk.MustNewDecFromStr(keeper.GetParams(ctx).TssSignerEmissionPercentage).Mul(blockRewards).TruncateInt()
	err := DistributeValidatorRewards(ctx, validatorRewards, keeper.GetBankKeeper(), keeper.GetFeeCollector())
	if err != nil {
		panic(err)
	}
	err = DistributeObserverRewards(ctx, observerRewards, keeper)
	if err != nil {
		panic(err)
	}
	err = DistributeTssRewards(ctx, tssSignerRewards, keeper.GetBankKeeper())
	if err != nil {
		panic(err)
	}
	types.EmitValidatorEmissions(ctx, bondFactor.String(), reservesFactor.String(),
		durationFactor.String(),
		validatorRewards.String(),
		observerRewards.String(),
		tssSignerRewards.String())
}

// DistributeValidatorRewards distributes the rewards to validators who signed the block .
// The block proposer gets a bonus reward
// This function uses the distribution module of cosmos-sdk , by directly sending funds to the feecollector.
func DistributeValidatorRewards(ctx sdk.Context, amount sdkmath.Int, bankKeeper types.BankKeeper, feeCollector string) error {
	coin := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount))
	return bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, feeCollector, coin)
}

// DistributeObserverRewards distributes the rewards to all observers who voted in any of the matured ballots
// The total rewards are distributed equally among all Successful votes
// NotVoted or Unsuccessful votes are slashed
// rewards given or slashed amounts are in azeta

func DistributeObserverRewards(ctx sdk.Context, amount sdkmath.Int, keeper keeper.Keeper) error {

	rewardsDistributer := map[string]int64{}
	totalRewardsUnits := int64(0)
	err := keeper.GetBankKeeper().SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedObserverRewardsPool, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount)))
	if err != nil {
		return err
	}
	ballotIdentifiers := keeper.GetObserverKeeper().GetMaturedBallotList(ctx)
	// do not distribute rewards if no ballots are matured, the rewards can accumulate in the undistributed pool
	if len(ballotIdentifiers) == 0 {
		return nil
	}
	for _, ballotIdentifier := range ballotIdentifiers {
		ballot, found := keeper.GetObserverKeeper().GetBallot(ctx, ballotIdentifier)
		if !found {
			continue
		}
		totalRewardsUnits += ballot.BuildRewardsDistribution(rewardsDistributer)
	}
	rewardPerUnit := sdkmath.ZeroInt()
	if totalRewardsUnits > 0 && amount.IsPositive() {
		rewardPerUnit = amount.Quo(sdk.NewInt(totalRewardsUnits))
	}

	sortedKeys := make([]string, 0, len(rewardsDistributer))
	for k := range rewardsDistributer {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var finalDistributionList []*types.ObserverEmission
	for _, key := range sortedKeys {
		observerAddress, err := sdk.AccAddressFromBech32(key)
		if err != nil {
			ctx.Logger().Error("Error while parsing observer address ", "error", err, "address", key)
			continue
		}
		// observerRewardUnits can be negative if the observer has been slashed
		// an observers earn 1 unit for a correct vote, and -1 unit for an incorrect vote
		observerRewardUnits := rewardsDistributer[key]

		if observerRewardUnits == 0 {
			finalDistributionList = append(finalDistributionList, &types.ObserverEmission{
				EmissionType:    types.EmissionType_Slash,
				ObserverAddress: observerAddress.String(),
				Amount:          sdkmath.ZeroInt(),
			})
			continue
		}
		if observerRewardUnits < 0 {
			slashAmount := keeper.GetParams(ctx).ObserverSlashAmount
			keeper.SlashObserverEmission(ctx, observerAddress.String(), slashAmount)
			finalDistributionList = append(finalDistributionList, &types.ObserverEmission{
				EmissionType:    types.EmissionType_Slash,
				ObserverAddress: observerAddress.String(),
				Amount:          slashAmount,
			})
			continue
		}
		// Defensive check
		if rewardPerUnit.GT(sdk.ZeroInt()) {
			rewardAmount := rewardPerUnit.Mul(sdkmath.NewInt(observerRewardUnits))
			keeper.AddObserverEmission(ctx, observerAddress.String(), rewardAmount)
			finalDistributionList = append(finalDistributionList, &types.ObserverEmission{
				EmissionType:    types.EmissionType_Rewards,
				ObserverAddress: observerAddress.String(),
				Amount:          rewardAmount,
			})
		}
	}
	types.EmitObserverEmissions(ctx, finalDistributionList)
	// TODO : Delete Ballots after distribution
	// https://github.com/zeta-chain/node/issues/942
	return nil
}

// DistributeTssRewards trasferes the allocated rewards to the Undistributed Tss Rewards Pool.
// This is done so that the reserves factor is properly calculated in the next block
func DistributeTssRewards(ctx sdk.Context, amount sdk.Int, bankKeeper types.BankKeeper) error {
	coin := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount))
	return bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedTssRewardsPool, coin)
}
