package keeper

import (
	"fmt"
	"strconv"
	"time"

	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeta-chain/node/common"
	"github.com/zeta-chain/node/x/crosschain/types"
	observertypes "github.com/zeta-chain/node/x/observer/types"
)

const (
	// RemainingFeesToStabilityPoolPercent is the percentage of remaining fees used to fund the gas stability pool
	RemainingFeesToStabilityPoolPercent = 95
)

// IterateAndUpdateCctxGasPrice iterates through all cctx and updates the gas price if pending for too long
func (k Keeper) IterateAndUpdateCctxGasPrice(ctx sdk.Context) error {
	// fetch the gas price increase flags or use default
	gasPriceIncreaseFlags := observertypes.DefaultGasPriceIncreaseFlags
	crosschainFlags, found := k.zetaObserverKeeper.GetCrosschainFlags(ctx)
	if found && crosschainFlags.GasPriceIncreaseFlags != nil {
		gasPriceIncreaseFlags = *crosschainFlags.GasPriceIncreaseFlags
	}

	// skip if haven't reached epoch end
	if ctx.BlockHeight()%gasPriceIncreaseFlags.EpochLength != 0 {
		return nil
	}

	// iterate all chains' pending cctx
	chains := common.DefaultChainsList()
	for _, chain := range chains {
		res, err := k.CctxAllPending(sdk.UnwrapSDKContext(ctx), &types.QueryAllCctxPendingRequest{
			ChainId: chain.ChainId,
		})
		if err != nil {
			return err
		}

		// iterate through all pending cctx
		for _, pendingCctx := range res.CrossChainTx {
			if pendingCctx != nil {
				_, _, err := k.CheckAndUpdateCctxGasPrice(ctx, *pendingCctx, gasPriceIncreaseFlags)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// CheckAndUpdateCctxGasPrice checks if the retry interval is reached and updates the gas price if so
// The function returns the gas price increase and the additional fees paid
func (k Keeper) CheckAndUpdateCctxGasPrice(
	ctx sdk.Context,
	cctx types.CrossChainTx,
	flags observertypes.GasPriceIncreaseFlags,
) (math.Uint, math.Uint, error) {
	// skip if gas price or gas limit is not set
	if cctx.GetCurrentOutTxParam().OutboundTxGasPrice == "" || cctx.GetCurrentOutTxParam().OutboundTxGasLimit == 0 {
		return math.ZeroUint(), math.ZeroUint(), nil
	}

	// skip if retry interval is not reached
	lastUpdated := time.Unix(cctx.CctxStatus.LastUpdateTimestamp, 0)
	if ctx.BlockTime().Before(lastUpdated.Add(flags.RetryInterval)) {
		return math.ZeroUint(), math.ZeroUint(), nil
	}

	// compute gas price increase
	chainID := cctx.GetCurrentOutTxParam().ReceiverChainId
	medianGasPrice, isFound := k.GetMedianGasPriceInUint(ctx, chainID)
	if !isFound {
		return math.ZeroUint(), math.ZeroUint(), cosmoserrors.Wrap(
			types.ErrUnableToGetGasPrice,
			fmt.Sprintf("cannot get gas price for chain %d", chainID),
		)
	}
	gasPriceIncrease := medianGasPrice.MulUint64(uint64(flags.GasPriceIncreasePercent)).QuoUint64(100)

	// withdraw additional fees from the gas stability pool
	gasLimit := math.NewUint(cctx.GetCurrentOutTxParam().OutboundTxGasLimit)
	additionalFees := gasLimit.Mul(gasPriceIncrease)
	if err := k.fungibleKeeper.WithdrawFromGasStabilityPool(ctx, chainID, additionalFees.BigInt()); err != nil {
		return math.ZeroUint(), math.ZeroUint(), cosmoserrors.Wrap(
			types.ErrNotEnoughFunds,
			fmt.Sprintf("cannot withdraw %s from gas stability pool", additionalFees.String()),
		)
	}

	// Increase the cctx value
	err := k.IncreaseCctxGasPrice(ctx, cctx, gasPriceIncrease)

	return gasPriceIncrease, additionalFees, err
}

// IncreaseCctxGasPrice increases the gas price associated with a CCTX and updates it in the store
func (k Keeper) IncreaseCctxGasPrice(ctx sdk.Context, cctx types.CrossChainTx, gasPriceIncrease math.Uint) error {
	currentGasPrice, err := strconv.ParseUint(cctx.GetCurrentOutTxParam().OutboundTxGasPrice, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse cctx gas price %s: %s", cctx.GetCurrentOutTxParam().OutboundTxGasPrice, err.Error())
	}

	// increase gas price and set last update timestamp
	cctx.GetCurrentOutTxParam().OutboundTxGasPrice = math.NewUint(currentGasPrice).Add(gasPriceIncrease).String()
	cctx.CctxStatus.LastUpdateTimestamp = ctx.BlockHeader().Time.Unix()
	k.SetCrossChainTx(ctx, cctx)

	return nil
}
