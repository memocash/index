package maint

import (
	"context"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
	"log"
)

var checkOrphansCmd = &cobra.Command{
	Use:   "check-orphans",
	Short: "Scan block heights for orphaned blocks",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		checkOrphans := maint.NewCheckOrphans(context.Background(), verbose)
		log.Println("Scanning block heights for orphans...")
		if err := checkOrphans.Check(); err != nil {
			log.Fatalf("error checking orphans; %v", err)
		}
		log.Printf("Done. Heights checked: %d, orphan blocks: %d, chain breaks: %d",
			checkOrphans.Total, checkOrphans.Orphans, checkOrphans.Breaks)
	},
}
