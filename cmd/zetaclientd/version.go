package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeta-chain/node/common"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "version description from git describe --tags",
	RunE:  Version,
}

func Version(_ *cobra.Command, _ []string) error {
	fmt.Printf(common.Version)
	return nil
}
