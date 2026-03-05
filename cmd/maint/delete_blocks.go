package maint

import (
	"context"
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var deleteBlocksCmd = &cobra.Command{
	Use:   "delete-blocks",
	Short: "Delete all block data from a given height onward",
	Run: func(c *cobra.Command, args []string) {
		start, _ := c.Flags().GetInt64(FlagStart)
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		dryRun, _ := c.Flags().GetBool(FlagDryRun)
		if start <= 0 {
			log.Fatalf("--start flag is required and must be positive")
		}
		deleteBlocks := maint.NewDeleteBlocks(context.Background(), start, verbose, dryRun)
		if dryRun {
			log.Printf("Dry run: scanning blocks from height %d onward...", start)
		} else {
			log.Printf("Deleting blocks from height %d onward...", start)
		}
		if err := deleteBlocks.Delete(); err != nil {
			log.Fatalf("error deleting blocks; %v", err)
		}
		log.Printf("Done. Blocks: %d, tx links: %d, duplicates: %d",
			deleteBlocks.Blocks, deleteBlocks.TxLinks, deleteBlocks.Duplicates)
	},
}
