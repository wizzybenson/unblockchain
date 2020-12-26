package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github/wizzybenson/unblockchain/database"
	"os"
)

const (
	flagFrom   = "from"
	flagTo     = "to"
	flagValue  = "value"
	flagReason = "reason"
)

func txCmd() *cobra.Command {
	var tx = &cobra.Command{
		Use:   "tx",
		Short: "Wrapper for tx actions like add and so on",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("incorrect usage")
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	tx.AddCommand(txAddCmd())

	return tx
}
func txAddCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add",
		Short: "command to add transaction to the database",
		Run: func(cmd *cobra.Command, args []string) {
			from, _ := cmd.Flags().GetString(flagFrom)
			to, _ := cmd.Flags().GetString(flagTo)
			value, _ := cmd.Flags().GetUint(flagValue)
			reason, _ := cmd.Flags().GetString(flagReason)

			fromAcc := database.NewAccount(from)
			toAcc := database.NewAccount(to)

			tx := database.NewTx(toAcc, fromAcc, value, reason)

			state, err := database.NewStateFromDisk()

			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			err = state.Add(tx)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			_, err = state.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

		},
	}

	cmd.Flags().String(flagFrom, "", "from what account to send tokens")
	cmd.MarkFlagRequired(flagFrom)
	cmd.Flags().String(flagTo, "", "to what account to send tokens")
	cmd.MarkFlagRequired(flagTo)
	cmd.Flags().Uint(flagValue, 0, "how much token to send")
	cmd.MarkFlagRequired(flagValue)
	cmd.Flags().String(flagReason, "", "possible value: 'reward'")
	return cmd
}
