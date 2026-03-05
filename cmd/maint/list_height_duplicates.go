package maint

import (
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var listHeightDuplicatesCmd = &cobra.Command{
	Use:   "list-height-duplicates",
	Short: "List all height duplicate entries",
	Run: func(c *cobra.Command, args []string) {
		listHeightDuplicates := new(maint.ListHeightDuplicates)
		log.Println("Listing height duplicates...")
		if err := listHeightDuplicates.List(c.Context()); err != nil {
			log.Fatalf("error listing height duplicates; %v", err)
		}
		log.Printf("Total height duplicates: %d\n", listHeightDuplicates.Total)
	},
}
