package maint

import (
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var doubleSpendCmd = &cobra.Command{
	Use:   "double-spends",
	Short: "Count double spends in output_inputs topic",
	Run: func(c *cobra.Command, args []string) {
		doubleSpends := new(maint.DoubleSpends)
		log.Println("Counting double spends...")
		if err := doubleSpends.Check(c.Context()); err != nil {
			log.Fatalf("error checking double spends; %v", err)
		}
		log.Printf("Total entries: %d\n", doubleSpends.TotalEntries)
		log.Printf("Double spend outputs: %d\n", doubleSpends.DoubleSpends)
	},
}
