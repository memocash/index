package maint

import (
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
	"log"
)

var populateOutputInputSingleCmd = &cobra.Command{
	Use:   "populate-output-input-single",
	Short: "populate-output-input-single",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populate := maint.NewPopulateOutputInputSingle(c.Context())
		log.Printf("Starting populate output input single...\n")
		if err := populate.Populate(restart); err != nil {
			log.Fatalf("error populate output input single; %v", err)
		}
		log.Printf("Populated output input single completed. Checked: %d, saved: %d, double spends: %d.\n",
			populate.Checked, populate.Saved, populate.DoubleSpends)
	},
}
