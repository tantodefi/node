package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateCoreParams{}, "observer/UpdateClientParams", nil)
	cdc.RegisterConcrete(&MsgUpdateCrosschainFlags{}, "crosschain/UpdateCrosschainFlags", nil)
	cdc.RegisterConcrete(&MsgUpdateKeygen{}, "crosschain/UpdateKeygen", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateCoreParams{},
		&MsgUpdateCrosschainFlags{},
		&MsgUpdateKeygen{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
