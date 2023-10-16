package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	eth "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/zeta-chain/node/common"

	fungibletypes "github.com/zeta-chain/node/x/fungible/types"
	zetaObserverTypes "github.com/zeta-chain/node/x/observer/types"
)

type StakingKeeper interface {
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
}

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(sdk.Context, types.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

type ZetaObserverKeeper interface {
	SetObserverMapper(ctx sdk.Context, om *zetaObserverTypes.ObserverMapper)
	GetObserverMapper(ctx sdk.Context, chain *common.Chain) (val zetaObserverTypes.ObserverMapper, found bool)
	GetAllObserverMappers(ctx sdk.Context) (mappers []*zetaObserverTypes.ObserverMapper)
	SetBallot(ctx sdk.Context, ballot *zetaObserverTypes.Ballot)
	GetBallot(ctx sdk.Context, index string) (val zetaObserverTypes.Ballot, found bool)
	GetAllBallots(ctx sdk.Context) (voters []*zetaObserverTypes.Ballot)
	GetParams(ctx sdk.Context) (params zetaObserverTypes.Params)
	GetCoreParamsByChainID(ctx sdk.Context, chainID int64) (params *zetaObserverTypes.CoreParams, found bool)
	GetNodeAccount(ctx sdk.Context, address string) (nodeAccount zetaObserverTypes.NodeAccount, found bool)
	GetAllNodeAccount(ctx sdk.Context) (nodeAccounts []zetaObserverTypes.NodeAccount)
	SetNodeAccount(ctx sdk.Context, nodeAccount zetaObserverTypes.NodeAccount)
	IsInboundEnabled(ctx sdk.Context) (found bool)
	GetCrosschainFlags(ctx sdk.Context) (val zetaObserverTypes.CrosschainFlags, found bool)
	GetKeygen(ctx sdk.Context) (val zetaObserverTypes.Keygen, found bool)
	SetKeygen(ctx sdk.Context, keygen zetaObserverTypes.Keygen)
	SetCrosschainFlags(ctx sdk.Context, crosschainFlags zetaObserverTypes.CrosschainFlags)
	SetLastObserverCount(ctx sdk.Context, lbc *zetaObserverTypes.LastObserverCount)
	AddVoteToBallot(ctx sdk.Context, ballot zetaObserverTypes.Ballot, address string, observationType zetaObserverTypes.VoteType) (zetaObserverTypes.Ballot, error)
	CheckIfFinalizingVote(ctx sdk.Context, ballot zetaObserverTypes.Ballot) (zetaObserverTypes.Ballot, bool)
	IsAuthorized(ctx sdk.Context, address string, chain *common.Chain) bool
	FindBallot(ctx sdk.Context, index string, chain *common.Chain, observationType zetaObserverTypes.ObservationType) (ballot zetaObserverTypes.Ballot, isNew bool, err error)
	AddBallotToList(ctx sdk.Context, ballot zetaObserverTypes.Ballot)
	GetBlockHeader(ctx sdk.Context, hash []byte) (val common.BlockHeader, found bool)
}

type FungibleKeeper interface {
	GetForeignCoins(ctx sdk.Context, zrc20Addr string) (val fungibletypes.ForeignCoins, found bool)
	GetAllForeignCoins(ctx sdk.Context) (list []fungibletypes.ForeignCoins)
	SetForeignCoins(ctx sdk.Context, foreignCoins fungibletypes.ForeignCoins)
	GetAllForeignCoinsForChain(ctx sdk.Context, foreignChainID int64) (list []fungibletypes.ForeignCoins)
	GetForeignCoinFromAsset(ctx sdk.Context, asset string, chainID int64) (fungibletypes.ForeignCoins, bool)
	GetSystemContract(ctx sdk.Context) (val fungibletypes.SystemContract, found bool)
	QuerySystemContractGasCoinZRC20(ctx sdk.Context, chainID *big.Int) (eth.Address, error)
	GetUniswapV2Router02Address(ctx sdk.Context) (eth.Address, error)
	QueryUniswapV2RouterGetZetaAmountsIn(ctx sdk.Context, amountOut *big.Int, outZRC4 eth.Address) (*big.Int, error)
	QueryUniswapV2RouterGetZRC4AmountsIn(ctx sdk.Context, amountOut *big.Int, inZRC4 eth.Address) (*big.Int, error)
	QueryUniswapV2RouterGetZRC4ToZRC4AmountsIn(ctx sdk.Context, amountOut *big.Int, inZRC4, outZRC4 eth.Address) (*big.Int, error)
	QueryGasLimit(ctx sdk.Context, contract eth.Address) (*big.Int, error)
	QueryProtocolFlatFee(ctx sdk.Context, contract eth.Address) (*big.Int, error)
	SetGasPrice(ctx sdk.Context, chainID *big.Int, gasPrice *big.Int) (uint64, error)
	DepositCoinZeta(ctx sdk.Context, to eth.Address, amount *big.Int) error
	DepositZRC20(
		ctx sdk.Context,
		contract eth.Address,
		to eth.Address,
		amount *big.Int,
	) (*evmtypes.MsgEthereumTxResponse, error)
	ZRC20DepositAndCallContract(
		ctx sdk.Context,
		from []byte,
		to eth.Address,
		amount *big.Int,
		senderChain *common.Chain,
		data []byte,
		coinType common.CoinType,
		asset string,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallUniswapV2RouterSwapExactTokensForTokens(
		ctx sdk.Context,
		sender eth.Address,
		to eth.Address,
		amountIn *big.Int,
		inZRC4,
		outZRC4 eth.Address,
		noEthereumTxEvent bool,
	) (ret []*big.Int, err error)
	CallUniswapV2RouterSwapExactETHForToken(
		ctx sdk.Context,
		sender eth.Address,
		to eth.Address,
		amountIn *big.Int,
		outZRC4 eth.Address,
		noEthereumTxEvent bool,
	) ([]*big.Int, error)
	CallZRC20Burn(ctx sdk.Context, sender eth.Address, zrc20address eth.Address, amount *big.Int, noEthereumTxEvent bool) error
	CallZRC20Approve(
		ctx sdk.Context,
		owner eth.Address,
		zrc20address eth.Address,
		spender eth.Address,
		amount *big.Int,
		noEthereumTxEvent bool,
	) error
	DeployZRC20Contract(
		ctx sdk.Context,
		name, symbol string,
		decimals uint8,
		chainID int64,
		coinType common.CoinType,
		erc20Contract string,
		gasLimit *big.Int,
	) (eth.Address, error)
	FundGasStabilityPool(ctx sdk.Context, chainID int64, amount *big.Int) error
	WithdrawFromGasStabilityPool(ctx sdk.Context, chainID int64, amount *big.Int) error
}
