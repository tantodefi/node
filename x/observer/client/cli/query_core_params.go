package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/zeta-chain/node/x/observer/types"
)

func CmdGetCoreParamsForChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-core-params [chain-id]",
		Short: "Query GetCoreParamsForChain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqChainID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetCoreParamsForChainRequest{
				ChainId: reqChainID,
			}
			res, err := queryClient.GetCoreParamsForChain(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetCoreParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-core-params",
		Short: "Query GetCoreParams",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetCoreParamsRequest{}
			res, err := queryClient.GetCoreParams(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
