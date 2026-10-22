package maint

import (
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var populateCashTokenCmd = &cobra.Command{
	Use:   "populate-cash-token",
	Short: "Backfill cash_token_output and addr_seen_tx rows for token-prefixed outputs",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populate := maint.NewPopulateCashToken(c.Context())
		log.Printf("Starting populate cash token...\n")
		if err := populate.Populate(restart); err != nil {
			log.Fatalf("error populate cash token; %v", err)
		}
		log.Printf("Populate cash token completed. Checked: %d, token outputs: %d, saved: %d.\n",
			populate.Checked, populate.Tokens, populate.Saved)
	},
}
