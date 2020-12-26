package main

import (
	"github.com/spf13/cobra"
	"fmt"
)

const (
	major = "0"
	minor = "1"
	fix = "1"
	verbal = "Tx add and balances list"
)

var versionCmd = &cobra.Command{
	Use: "version",
	Short: "Version Description",
	Run: func (cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("Version %s %s %s-beta %s", major, minor, fix, verbal))
	},
}
