package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	/*s := []string{"a", "b", "c", "d", "e"}
	mempool := make([]string, len(s))
	copy(mempool, s)
	for i := 0; i < len(mempool); i++ {
		s = append(s[:i], s[i+1:]...)
	}*/
	tub := &cobra.Command{
		Use:   "tub",
		Short: "The unblockchain bar CLI",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	tub.AddCommand(versionCmd)
	tub.AddCommand(balancesCmd())
	tub.AddCommand(txCmd())
	if err := tub.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
