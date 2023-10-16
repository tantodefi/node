package zetaclient

import (
	"github.com/zeta-chain/node/common"
)

// Modify to update this from the core later
func GetSupportedChains() []*common.Chain {
	return common.DefaultChainsList()
}
