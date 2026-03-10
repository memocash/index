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
		doubleSpends, _ := c.Flags().GetBool(FlagDoubleSpends)
		listHeightDuplicates := maint.NewListHeightDuplicates(doubleSpends)
		log.Println("Listing height duplicates...")
		if err := listHeightDuplicates.List(c.Context()); err != nil {
			log.Fatalf("error listing height duplicates; %v", err)
		}
		if doubleSpends {
			log.Printf("Total height duplicates: %d, double spends: %d\n", listHeightDuplicates.Total, listHeightDuplicates.DoubleSpends)
		} else {
			log.Printf("Total height duplicates: %d\n", listHeightDuplicates.Total)
		}
	},
}
