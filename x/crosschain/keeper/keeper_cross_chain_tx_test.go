package keeper

import (
	"fmt"
	"math/rand"
	"testing"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/zeta-chain/node/common"
	"github.com/zeta-chain/node/x/crosschain/types"
)

func createNCctxWithStatus(keeper *Keeper, ctx sdk.Context, n int, status types.CctxStatus) []types.CrossChainTx {
	items := make([]types.CrossChainTx, n)
	for i := range items {
		items[i].Creator = "any"
		items[i].Index = fmt.Sprintf("%d-%d", i, status)
		items[i].CctxStatus = &types.Status{
			Status:              status,
			StatusMessage:       "",
			LastUpdateTimestamp: 0,
		}
		items[i].ZetaFees = math.OneUint()
		items[i].InboundTxParams = &types.InboundTxParams{InboundTxObservedHash: fmt.Sprintf("%d", i), Amount: math.OneUint()}

		keeper.SetCctxAndNonceToCctxAndInTxHashToCctx(ctx, items[i])
	}
	return items
}

// Keeper Tests
func createNCctx(keeper *Keeper, ctx sdk.Context, n int) []types.CrossChainTx {
	items := make([]types.CrossChainTx, n)
	for i := range items {
		items[i].Creator = "any"
		items[i].InboundTxParams = &types.InboundTxParams{
			Sender:                          fmt.Sprintf("%d", i),
			SenderChainId:                   int64(i),
			TxOrigin:                        fmt.Sprintf("%d", i),
			Asset:                           fmt.Sprintf("%d", i),
			CoinType:                        common.CoinType_Zeta,
			InboundTxObservedHash:           fmt.Sprintf("%d", i),
			InboundTxObservedExternalHeight: uint64(i),
			InboundTxFinalizedZetaHeight:    uint64(i),
		}
		items[i].OutboundTxParams = []*types.OutboundTxParams{{
			Receiver:                         fmt.Sprintf("%d", i),
			ReceiverChainId:                  int64(i),
			OutboundTxHash:                   fmt.Sprintf("%d", i),
			OutboundTxTssNonce:               uint64(i),
			OutboundTxGasLimit:               uint64(i),
			OutboundTxGasPrice:               fmt.Sprintf("%d", i),
			OutboundTxBallotIndex:            fmt.Sprintf("%d", i),
			OutboundTxObservedExternalHeight: uint64(i),
			CoinType:                         0,
		}}
		items[i].CctxStatus = &types.Status{
			Status:              types.CctxStatus_PendingInbound,
			StatusMessage:       "any",
			LastUpdateTimestamp: 0,
		}
		items[i].InboundTxParams.Amount = math.OneUint()

		items[i].ZetaFees = math.OneUint()
		items[i].Index = fmt.Sprintf("%d", i)
		keeper.SetCctxAndNonceToCctxAndInTxHashToCctx(ctx, items[i])
	}
	return items
}

func TestSends(t *testing.T) {
	sendsTest := []struct {
		TestName        string
		PendingInbound  int
		PendingOutbound int
		OutboundMined   int
		Confirmed       int
		PendingRevert   int
		Reverted        int
		Aborted         int
	}{
		{
			TestName:        "test pending",
			PendingInbound:  10,
			PendingOutbound: 10,
			Confirmed:       10,
			PendingRevert:   10,
			Aborted:         10,
			OutboundMined:   10,
			Reverted:        10,
		},
		{
			TestName:        "test pending random",
			PendingInbound:  rand.Intn(300-10) + 10,
			PendingOutbound: rand.Intn(300-10) + 10,
			Confirmed:       rand.Intn(300-10) + 10,
			PendingRevert:   rand.Intn(300-10) + 10,
			Aborted:         rand.Intn(300-10) + 10,
			OutboundMined:   rand.Intn(300-10) + 10,
			Reverted:        rand.Intn(300-10) + 10,
		},
	}
	for _, tt := range sendsTest {
		tt := tt
		t.Run(tt.TestName, func(t *testing.T) {
			keeper, ctx := setupKeeper(t)
			var sends []types.CrossChainTx
			sends = append(sends, createNCctxWithStatus(keeper, ctx, tt.PendingInbound, types.CctxStatus_PendingInbound)...)
			sends = append(sends, createNCctxWithStatus(keeper, ctx, tt.PendingOutbound, types.CctxStatus_PendingOutbound)...)
			sends = append(sends, createNCctxWithStatus(keeper, ctx, tt.PendingRevert, types.CctxStatus_PendingRevert)...)
			sends = append(sends, createNCctxWithStatus(keeper, ctx, tt.Aborted, types.CctxStatus_Aborted)...)
			sends = append(sends, createNCctxWithStatus(keeper, ctx, tt.OutboundMined, types.CctxStatus_OutboundMined)...)
			sends = append(sends, createNCctxWithStatus(keeper, ctx, tt.Reverted, types.CctxStatus_Reverted)...)
			//assert.Equal(t, tt.PendingOutbound, len(keeper.GetAllCctxByStatuses(ctx, []types.CctxStatus{types.CctxStatus_PendingOutbound})))
			//assert.Equal(t, tt.PendingInbound, len(keeper.GetAllCctxByStatuses(ctx, []types.CctxStatus{types.CctxStatus_PendingInbound})))
			//assert.Equal(t, tt.PendingOutbound+tt.PendingRevert, len(keeper.GetAllCctxByStatuses(ctx, []types.CctxStatus{types.CctxStatus_PendingOutbound, types.CctxStatus_PendingRevert})))
			assert.Equal(t, len(sends), len(keeper.GetAllCrossChainTx(ctx)))
			for _, s := range sends {
				send, found := keeper.GetCrossChainTx(ctx, s.Index)
				assert.True(t, found)
				assert.Equal(t, s, send)
			}

		})
	}
}

func TestSendGetAll(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	items := createNCctx(keeper, ctx, 10)
	cctx := keeper.GetAllCrossChainTx(ctx)
	c := make([]types.CrossChainTx, len(cctx))
	for i, val := range cctx {
		c[i] = val
	}
	assert.Equal(t, items, c)
}

// Querier Tests

func TestSendQuerySingle(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNCctx(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetCctxRequest
		response *types.QueryGetCctxResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetCctxRequest{Index: msgs[0].Index},
			response: &types.QueryGetCctxResponse{CrossChainTx: &msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetCctxRequest{Index: msgs[1].Index},
			response: &types.QueryGetCctxResponse{CrossChainTx: &msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetCctxRequest{Index: "missing"},
			err:     status.Error(codes.InvalidArgument, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Cctx(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.Equal(t, tc.response, response)
			}
		})
	}
}

func TestSendQueryPaginated(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNCctx(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllCctxRequest {
		return &types.QueryAllCctxRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.CctxAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			for j := i; j < len(msgs) && j < i+step; j++ {
				assert.Equal(t, &msgs[j], resp.CrossChainTx[j-i])
			}
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.CctxAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			for j := i; j < len(msgs) && j < i+step; j++ {
				assert.Equal(t, &msgs[j], resp.CrossChainTx[j-i])
			}
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.CctxAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.CctxAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
