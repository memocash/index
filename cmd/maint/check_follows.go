package maint

import (
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
	"log"
)

var checkFollowsCmd = &cobra.Command{
	Use:   "check-follows",
	Short: "check-follows",
	Run: func(c *cobra.Command, args []string) {
		deleteItems, _ := c.Flags().GetBool(FlagDelete)
		checkFollows := maint.NewCheckFollows(deleteItems)
		log.Printf("Starting check follows (delete flag: %t)...\n", deleteItems)
		if err := checkFollows.Check(); err != nil {
			log.Fatalf("error maint check follows; %v", err)
		}
		log.Printf("Checked follows: %d, bad: %d\n", checkFollows.Processed, checkFollows.BadFollows)
	},
}
