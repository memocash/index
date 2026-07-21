package maint

import (
	"context"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
	"log"
)

var populateLinkRequestParentsCmd = &cobra.Command{
	Use:   "populate-link-request-parents",
	Short: "Save a link request parent index item for each existing memo addr link request",
	Run: func(c *cobra.Command, args []string) {
		populateLinkRequestParent := maint.NewPopulateLinkRequestParent(context.Background())
		log.Printf("Starting populate link request parents...\n")
		if err := populateLinkRequestParent.Populate(); err != nil {
			log.Fatalf("error populate link request parents; %v", err)
		}
		log.Printf("Populate link request parents completed. Requests: %d, parents: %d, skipped: %d.\n",
			populateLinkRequestParent.Requests, populateLinkRequestParent.Parents, populateLinkRequestParent.Skipped)
	},
}
