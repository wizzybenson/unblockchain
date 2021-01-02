package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github/wizzybenson/unblockchain/database"
	"os"
)

func balancesCmd() *cobra.Command {
	var balancesCmd = &cobra.Command{
		Use:   "balances",
		Short: "Wrapper to manage balances (list and other commands",
		PreRunE: func(cmd *cobra.Command, args []string) error{
			return fmt.Errorf("incorrect usage")
		},
		Run: func(cm *cobra.Command, args []string) {

		},
	}
	balancesCmd.AddCommand(balancesListCmd)
	return balancesCmd
}

var balancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists balances",
	Long:  "Lists balances in the state component",
	Run: func(cmd *cobra.Command, args []string) {
		state, err := database.NewStateFromDisk()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer state.Close()
		fmt.Printf("Account balances at %x\n", state.LatestBlockHash())
		fmt.Println("-----------------------")
		fmt.Println("")

		for account, balance := range state.Balances {
			fmt.Println(fmt.Sprintf("%s: %d", account, balance))
		}
	},
}
