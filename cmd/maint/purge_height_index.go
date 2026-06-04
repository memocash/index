package maint

import (
	"context"
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var purgeHeightIndexCmd = &cobra.Command{
	Use:   "purge-height-index",
	Short: "Delete height-index entries (HeightBlock + BlockHeight) at/above a start height, preserving block and tx data",
	Run: func(c *cobra.Command, args []string) {
		start, _ := c.Flags().GetInt64(FlagStart)
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		dryRun, _ := c.Flags().GetBool(FlagDryRun)
		verify, _ := c.Flags().GetBool(FlagVerify)
		if start <= 0 {
			log.Fatalf("--start flag is required and must be positive")
		}
		purge := maint.NewPurgeHeightIndex(context.Background(), start, verbose, dryRun, verify)
		if dryRun {
			log.Printf("Dry run: scanning height-index entries from height %d onward...", start)
		} else {
			log.Printf("Purging height-index entries from height %d onward...", start)
		}
		if err := purge.Purge(); err != nil {
			log.Fatalf("error purging height index; %v", err)
		}
		log.Printf("Done. Height mappings purged: %d, skipped (no lower copy): %d, height duplicates: %d",
			purge.HeightBlocks, purge.Skipped, purge.Duplicates)
	},
}
