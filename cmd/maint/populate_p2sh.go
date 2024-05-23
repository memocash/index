package maint

import (
	"context"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
	"log"
)

var populateP2shCmd = &cobra.Command{
	Use:   "populate-p2sh",
	Short: "populate-p2sh",
	Run: func(c *cobra.Command, args []string) {
		var startHeight int64
		if len(args) > 0 {
			startHeight = jutil.GetInt64FromString(args[0])
		}
		populateP2sh := maint.NewPopulateP2sh(context.Background())
		log.Printf("Starting populate p2sh...\n")
		if err := populateP2sh.Populate(startHeight); err != nil {
			log.Fatalf("error populate p2sh; %v", err)
		}
		log.Printf("Populated p2sh, blocks processed: %d\n", populateP2sh.BlocksProcessed)
	},
}

var populateP2shDirectCmd = &cobra.Command{
	Use:   "populate-p2sh-direct",
	Short: "populate-p2sh-direct",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populateP2sh := maint.NewPopulateP2shDirect(context.Background())
		log.Printf("Starting populate p2sh...\n")
		if err := populateP2sh.Populate(restart); err != nil {
			log.Fatalf("error populate p2sh; %v", err)
		}
		log.Printf("Populated p2sh direct completed. Checked: %d, saved: %d.\n", populateP2sh.Checked, populateP2sh.Saved)
	},
}
