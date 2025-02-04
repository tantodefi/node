package querytests

import (
	"fmt"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/zeta-chain/node/x/observer/client/cli"
	observerTypes "github.com/zeta-chain/node/x/observer/types"
	"google.golang.org/grpc/status"
)

func (s *CliTestSuite) TestShowKeygen() {
	ctx := s.network.Validators[0].ClientCtx
	obj := s.observerState.Keygen
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		args []string
		err  error
		obj  *observerTypes.Keygen
	}{
		{
			desc: "get",
			args: common,
			obj:  obj,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			var args []string
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowKeygen(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				s.Require().True(ok)
				s.Require().ErrorIs(stat.Err(), tc.err)
			} else {
				s.Require().NoError(err)
				var resp observerTypes.QueryGetKeygenResponse
				s.Require().NoError(s.network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				s.Require().NotNil(resp.Keygen)
				s.Require().Equal(tc.obj, resp.Keygen)
			}
		})
	}
}
