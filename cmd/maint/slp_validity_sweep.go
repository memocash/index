package maint

import (
	"context"
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var slpValiditySweepCmd = &cobra.Command{
	Use:   "slp-validity-sweep",
	Short: "Validate SLP txs by height, writing slp_validity verdicts",
	Long: "Walks blocks by height and validates every SLP tx to a fixpoint within each block. " +
		"Serves as both the historical backfill and the safety net for txs the live save path " +
		"leaves pending. Resumes from its saved cursor unless --start is given; verdicts are " +
		"final so re-running is idempotent.",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		start, _ := c.Flags().GetInt64(FlagStart)
		end, _ := c.Flags().GetInt64(FlagEnd)
		sweep := maint.NewSlpValiditySweep(context.Background(), verbose, start, end)
		if err := sweep.Run(); err != nil {
			log.Fatalf("error running slp validity sweep; %v", err)
		}
	},
}
