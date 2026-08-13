package maint

import (
	"context"
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var slpValiditySweepCmd = &cobra.Command{
	Use:   "slp-validity-sweep",
	Short: "Validate undecided SLP txs from the index's own datasets",
	Long: "Iterates the slp genesis/mint/send topics and validates any tx without a verdict — " +
		"the routine safety net for txs the live save path left pending. For txs the live path " +
		"never transcribed (historical backfill / deep audit), use slp-validity-backfill, which " +
		"scans the raw chain tx outputs. Verdicts are final so re-running is idempotent.",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		sweep := maint.NewSlpValiditySweep(context.Background(), verbose)
		if err := sweep.Run(); err != nil {
			log.Fatalf("error running slp validity sweep; %v", err)
		}
	},
}
