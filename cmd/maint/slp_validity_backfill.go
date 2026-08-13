package maint

import (
	"context"
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var slpValidityBackfillCmd = &cobra.Command{
	Use:   "slp-validity-backfill",
	Short: "Validate all undecided SLP txs once, in block order",
	Long: "Finds every SLP candidate with a server-side filtered scan of chain tx outputs, then " +
		"validates the undecided ones in block-height order (topologically sorted within each " +
		"chunk, since CTOR blocks order txs by txid): parents always decide before or alongside " +
		"their children, so each tx is validated exactly once instead of cascading one generation " +
		"per round. Unmined candidates are cascaded at the end. Verdicts are final, so re-running " +
		"is idempotent and a restart simply re-scans; already-decided txs drop out up front.",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		backfill := maint.NewSlpValidityBackfill(context.Background(), verbose)
		if err := backfill.Run(); err != nil {
			log.Fatalf("error running slp validity backfill; %v", err)
		}
	},
}
