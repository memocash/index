package maint

import (
	"log"

	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/spf13/cobra"
)

var setBlockHeightCmd = &cobra.Command{
	Use:   "set-block-height",
	Short: "Manually set the sync status block height",
	Run: func(c *cobra.Command, args []string) {
		height, _ := c.Flags().GetInt64(FlagHeight)
		if height < 0 {
			log.Fatalf("--height must be non-negative")
		}
		if err := db.Save([]db.Object{&item.SyncStatus{
			Name:   item.SyncStatusBlockHeight,
			Height: height,
		}}); err != nil {
			log.Fatalf("error setting sync status block height; %v", err)
		}
		log.Printf("Set sync status block height to %d", height)
	},
}
