package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeta-chain/node/testutil/sample"
	"github.com/zeta-chain/node/x/fungible/types"
)

func TestMsgUpdateContractBytecode_ValidateBasic(t *testing.T) {
	tt := []struct {
		name      string
		msg       types.MsgUpdateContractBytecode
		wantError bool
	}{
		{
			name: "valid",
			msg: types.MsgUpdateContractBytecode{
				Creator:            sample.AccAddress(),
				ContractAddress:    sample.EthAddress().Hex(),
				NewBytecodeAddress: sample.EthAddress().Hex(),
			},
			wantError: false,
		},
		{
			name: "invalid creator",
			msg: types.MsgUpdateContractBytecode{
				Creator:            "invalid",
				ContractAddress:    sample.EthAddress().Hex(),
				NewBytecodeAddress: sample.EthAddress().Hex(),
			},
			wantError: true,
		},
		{
			name: "invalid contract address",
			msg: types.MsgUpdateContractBytecode{
				Creator:            sample.AccAddress(),
				ContractAddress:    "invalid",
				NewBytecodeAddress: sample.EthAddress().Hex(),
			},
			wantError: true,
		},
		{
			name: "invalid bytecode address",
			msg: types.MsgUpdateContractBytecode{
				Creator:            sample.AccAddress(),
				ContractAddress:    sample.EthAddress().Hex(),
				NewBytecodeAddress: "invalid",
			},
			wantError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
