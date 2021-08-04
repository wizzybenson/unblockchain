package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github/wizzybenson/unblockchain/database"
	"github/wizzybenson/unblockchain/node"
	"github/wizzybenson/unblockchain/wallet"
	"time"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "The Unblockchain Bar migration command",
		Run: func(cmd *cobra.Command, args []string) {
			ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)
			miner, _ := cmd.Flags().GetString(flagMiner)

			thanos := database.NewAccount(wallet.ThanosAccount)
			maw := database.NewAccount(wallet.MawAccount)
			proxima := database.NewAccount(wallet.ProximaAccount)

			peer := node.NewPeerNode(
				"127.0.0.1",
				8086,
				true,
				thanos,
				false,
			)
			n := node.New(getDataDirFromCmd(cmd), ip, port, database.NewAccount(miner), peer)

			n.AddPendingTX(database.NewTx(thanos, thanos, 3, ""), peer)
			n.AddPendingTX(database.NewTx(thanos, thanos, 700, ""), peer)
			n.AddPendingTX(database.NewTx(maw, thanos, 2000, ""), peer)
			n.AddPendingTX(database.NewTx(thanos, thanos, 100, ""), peer)
			n.AddPendingTX(database.NewTx(thanos, maw, 1, ""), peer)
			n.AddPendingTX(database.NewTx(proxima, maw, 1000, ""), peer)
			n.AddPendingTX(database.NewTx(thanos, maw, 50, ""), peer)

			ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*15)

			go func() {
				ticker := time.NewTicker(time.Second * 10)

				for {
					select {
					case <-ticker.C:
						if !n.LatestBlockHash().IsEmpty() {
							closeNode()
							return
						}
					}
				}
			}()

			err := n.Run(ctx)
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	addDefaultRequiredCmds(migrateCmd)
	migrateCmd.Flags().String(flagIP, node.DefaultIp, "expose IP for communication with peers")
	migrateCmd.Flags().Uint64(flagPort, node.DefaultHttpPort, "expose HTTP port for communication with peers")
	migrateCmd.Flags().String(flagMiner, node.DefaultMiner, "Name of the node owner")

	return migrateCmd
}
