package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github/wizzybenson/unblockchain/fs"
	"os"
)
const flagDatadir = "datadir"
const flagPort = "port"
const flagIP ="ip"
const flagMiner = "miner"
const flagKeystoreFile = "keystore"
const flagBootstrapIp = "bootstrap-ip"
const flagBootstrapAcc = "boostrap-account"
const flagBootstrapPort = "bootstrap-port"

func main() {

	tub := &cobra.Command{
		Use:   "tub",
		Short: "The unblockchain bar CLI",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	tub.AddCommand(versionCmd)
	tub.AddCommand(balancesCmd())
	tub.AddCommand(runCmd())
	tub.AddCommand(migrateCmd())
	tub.AddCommand(walletCmd())
	if err := tub.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredCmds(cmd *cobra.Command) {
	cmd.Flags().String(flagDatadir, "", "Absolute path of where to store blockchain data")
	cmd.MarkFlagRequired(flagDatadir)
}

func addKeystoreFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagKeystoreFile, "", "Absolute path of where the wallet keystore is stored")
	cmd.MarkFlagRequired(flagKeystoreFile)
}

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDatadir)

	return fs.ExpandPath(dataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}