package maint

import (
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

const (
	FlagStart = "start"
	FlagEnd   = "end"
)

var backfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Process full blocks in a height range through the saver pipeline",
	Run: func(c *cobra.Command, args []string) {
		start, _ := c.Flags().GetInt64(FlagStart)
		end, _ := c.Flags().GetInt64(FlagEnd)
		if start <= 0 || end <= 0 {
			log.Fatalf("both --start and --end flags are required and must be positive")
		}
		if start > end {
			log.Fatalf("--start (%d) must be less than or equal to --end (%d)", start, end)
		}
		backfill := maint.NewBackfill(start, end)
		log.Printf("Starting backfill from %d to %d...\n", start, end)
		if err := backfill.Run(); err != nil {
			log.Fatalf("error running backfill; %v", err)
		}
	},
}
