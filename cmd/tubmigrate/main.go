package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github/wizzybenson/unblockchain/database"
	"os"
	"time"
)

func main() {
	var tubMigrateCmd = &cobra.Command{
		Use:   "tubmigrate",
		Short: "The Blockchain Bar migration command",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()
			block0 := database.NewBlock(
				database.Hash{},
				uint64(time.Now().Unix()),
				[]database.Tx{database.NewTx("thanos", "thanos", 3, ""),
					database.NewTx("thanos", "thanos", 700, "reward")},
			)

			state.AddBlock(block0)
			block0Hash, err := state.Persist()

			block1 := database.NewBlock(
				block0Hash,
				uint64(time.Now().Unix()),
				[]database.Tx{database.NewTx("maw", "thanos", 2000, ""),
					database.NewTx("thanos", "thanos", 100, "reward"),
					database.NewTx("thanos", "maw", 1, ""),
					database.NewTx("proxima", "maw", 1000, ""),
					database.NewTx("thanos", "maw", 50, ""),
					database.NewTx("thanos", "thanos", 100, "reward"),
					database.NewTx("thanos", "thanos", 100, "reward"),
				},
			)

			state.AddBlock(block1)
			state.Persist()
		},
	}

	if err := tubMigrateCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}


}
