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
		"the routine safety net for txs the live save path left pending. With --audit, instead " +
		"scans every chain tx output for an SLP lock script, transcribing and validating anything " +
		"the live path missed (historical backfill / deep audit); an interrupted audit resumes " +
		"from its per-shard cursor. The cursor survives an slp-topic wipe, so pass --fresh to " +
		"ignore it and start from the beginning (required after a wipe-and-repopulate, or the " +
		"audit silently skips everything before the old cursor). Verdicts are final so " +
		"re-running either mode is idempotent.",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		audit, _ := c.Flags().GetBool(FlagAudit)
		fresh, _ := c.Flags().GetBool(FlagFresh)
		sweep := maint.NewSlpValiditySweep(context.Background(), verbose, audit)
		sweep.Fresh = fresh
		if err := sweep.Run(); err != nil {
			log.Fatalf("error running slp validity sweep; %v", err)
		}
	},
}
