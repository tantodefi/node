package keeper

import (
	"errors"
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/zeta-chain/node/common"
	"github.com/zeta-chain/node/x/crosschain/types"
	zetaObserverTypes "github.com/zeta-chain/node/x/observer/types"
)

// UpdateNonce sets the CCTX outbound nonce to the next nonce, and updates the nonce of blockchain state.
// It also updates the PendingNonces that is used to track the unfulfilled outbound txs.
func (k Keeper) UpdateNonce(ctx sdk.Context, receiveChainID int64, cctx *types.CrossChainTx) error {
	chain := k.zetaObserverKeeper.GetParams(ctx).GetChainFromChainID(receiveChainID)
	if chain == nil {
		return zetaObserverTypes.ErrSupportedChains
	}

	nonce, found := k.GetChainNonces(ctx, chain.ChainName.String())
	if !found {
		return cosmoserrors.Wrap(types.ErrCannotFindReceiverNonce, fmt.Sprintf("Chain(%s) | Identifiers : %s ", chain.ChainName.String(), cctx.LogIdentifierForCCTX()))
	}

	// SET nonce
	cctx.GetCurrentOutTxParam().OutboundTxTssNonce = nonce.Nonce
	tss, found := k.GetTSS(ctx)
	if !found {
		return cosmoserrors.Wrap(types.ErrCannotFindTSSKeys, fmt.Sprintf("Chain(%s) | Identifiers : %s ", chain.ChainName.String(), cctx.LogIdentifierForCCTX()))
	}

	p, found := k.GetPendingNonces(ctx, tss.TssPubkey, receiveChainID)
	if !found {
		return cosmoserrors.Wrap(types.ErrCannotFindPendingNonces, fmt.Sprintf("chain_id %d, nonce %d", receiveChainID, nonce.Nonce))
	}

	// #nosec G701 always in range
	if p.NonceHigh != int64(nonce.Nonce) {
		return cosmoserrors.Wrap(types.ErrNonceMismatch, fmt.Sprintf("chain_id %d, high nonce %d, current nonce %d", receiveChainID, p.NonceHigh, nonce.Nonce))
	}

	nonce.Nonce++
	p.NonceHigh++
	k.SetChainNonces(ctx, nonce)
	k.SetPendingNonces(ctx, p)
	return nil
}

// RefundAmountOnZetaChain refunds the amount of the cctx on ZetaChain in case of aborted cctx
// NOTE: GetCurrentOutTxParam should contain the last up to date cctx amount
func (k Keeper) RefundAmountOnZetaChain(ctx sdk.Context, cctx types.CrossChainTx, inputAmount math.Uint) error {
	// preliminary checks
	if cctx.InboundTxParams.CoinType != common.CoinType_ERC20 {
		return errors.New("unsupported coin type for refund on ZetaChain")
	}
	if !common.IsEVMChain(cctx.InboundTxParams.SenderChainId) {
		return errors.New("only EVM chains are supported for refund on ZetaChain")
	}
	sender := ethcommon.HexToAddress(cctx.InboundTxParams.Sender)
	if sender == (ethcommon.Address{}) {
		return errors.New("invalid sender address")
	}
	if inputAmount.IsNil() || inputAmount.IsZero() {
		return errors.New("no amount to refund")
	}

	// get address of the zrc20
	fc, found := k.fungibleKeeper.GetForeignCoinFromAsset(ctx, cctx.InboundTxParams.Asset, cctx.InboundTxParams.SenderChainId)
	if !found {
		return fmt.Errorf("asset %s zrc not found", cctx.InboundTxParams.Asset)
	}
	zrc20 := ethcommon.HexToAddress(fc.Zrc20ContractAddress)
	if zrc20 == (ethcommon.Address{}) {
		return fmt.Errorf("asset %s invalid zrc address", cctx.InboundTxParams.Asset)
	}

	// deposit the amount to the sender
	if _, err := k.fungibleKeeper.DepositZRC20(ctx, zrc20, sender, inputAmount.BigInt()); err != nil {
		return errors.New("failed to deposit zrc20 on ZetaChain" + err.Error())
	}

	return nil
}
