package main

import (
	"github.com/spf13/cobra"
	"fmt"
)

const (
	major = "0"
	minor = "7"
	fix = "2"
	verbal = "Sync"
)

var versionCmd = &cobra.Command{
	Use: "version",
	Short: "Version Description",
	Run: func (cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("Version %s %s %s-beta %s", major, minor, fix, verbal))
	},
}
