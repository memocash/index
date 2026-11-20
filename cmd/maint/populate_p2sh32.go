package maint

import (
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var populateP2sh32OutputsCmd = &cobra.Command{
	Use:   "populate-p2sh32-outputs",
	Short: "Backfill addr_seen_tx rows for P2SH32 outputs",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populate := maint.NewPopulateP2sh32(c.Context(), false)
		log.Printf("Starting populate p2sh32 outputs...\n")
		if err := populate.Populate(restart); err != nil {
			log.Fatalf("error populate p2sh32 outputs; %v", err)
		}
		log.Printf("Populate p2sh32 outputs completed. Checked: %d, p2sh32 outputs: %d, saved: %d.\n",
			populate.Checked, populate.Found, populate.Saved)
	},
}

var populateP2shLinksCmd = &cobra.Command{
	Use:   "populate-p2sh-links",
	Short: "Backfill addr_p2sh_link rows (P2SH20 <-> P2SH32 of the same redeem script) from spent inputs",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populate := maint.NewPopulateP2sh32(c.Context(), true)
		log.Printf("Starting populate p2sh links...\n")
		if err := populate.Populate(restart); err != nil {
			log.Fatalf("error populate p2sh links; %v", err)
		}
		log.Printf("Populate p2sh links completed. Checked: %d, p2sh inputs: %d, saved: %d.\n",
			populate.Checked, populate.Found, populate.Saved)
	},
}
