//go:build TESTNET

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/zeta-chain/node/common"
)

// This is for privnet/testnet only
func (k Keeper) BlockOneDeploySystemContracts(goCtx context.Context) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// setup uniswap v2 factory
	uniswapV2Factory, err := k.DeployUniswapV2Factory(ctx)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to DeployUniswapV2Factory")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("UniswapV2Factory", uniswapV2Factory.String()),
		),
	)

	// setup WZETA contract
	wzeta, err := k.DeployWZETA(ctx)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to DeployWZetaContract")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("DeployWZetaContract", wzeta.String()),
		),
	)

	router, err := k.DeployUniswapV2Router02(ctx, uniswapV2Factory, wzeta)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to DeployUniswapV2Router02")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("DeployUniswapV2Router02", router.String()),
		),
	)

	connector, err := k.DeployConnectorZEVM(ctx, wzeta)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to DeployConnectorZEVM")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("DeployConnectorZEVM", connector.String()),
		),
	)
	ctx.Logger().Info("Deployed Connector ZEVM at " + connector.String())

	SystemContractAddress, err := k.DeploySystemContract(ctx, wzeta, uniswapV2Factory, router)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to SystemContractAddress")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("SystemContractAddress", SystemContractAddress.String()),
		),
	)

	// set the system contract
	system, _ := k.GetSystemContract(ctx)
	system.SystemContract = SystemContractAddress.String()
	// FIXME: remove unnecessary SetGasPrice and setupChainGasCoinAndPool
	k.SetSystemContract(ctx, system)
	//err = k.SetGasPrice(ctx, big.NewInt(1337), big.NewInt(1))
	if err != nil {
		return err
	}
	_, err = k.SetupChainGasCoinAndPool(ctx, common.GoerliChain().ChainId, "ETH", "gETH", 18, nil)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to setupChainGasCoinAndPool")
	}

	_, err = k.SetupChainGasCoinAndPool(ctx, common.BscTestnetChain().ChainId, "BNB", "tBNB", 18, nil)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to setupChainGasCoinAndPool")
	}
	_, err = k.SetupChainGasCoinAndPool(ctx, common.MumbaiChain().ChainId, "MATIC", "tMATIC", 18, nil)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to setupChainGasCoinAndPool")
	}
	_, err = k.SetupChainGasCoinAndPool(ctx, common.BtcTestNetChain().ChainId, "BTC", "tBTC", 8, nil)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to setupChainGasCoinAndPool")
	}

	return nil
}

func (k Keeper) TestUpdateSystemContractAddress(goCtx context.Context) error {
	return nil
}

func (k Keeper) TestUpdateZRC20WithdrawFee(goCtx context.Context) error {
	return nil
}
