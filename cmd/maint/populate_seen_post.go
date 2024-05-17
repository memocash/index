package maint

import (
	"context"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
	"log"
)

var populateSeenPostsCmd = &cobra.Command{
	Use: "populate-seen-posts",
	Run: func(c *cobra.Command, args []string) {
		populateSeenPost := maint.NewPopulateSeenPost(context.Background())
		log.Printf("Starting populate seen posts...\n")
		if err := populateSeenPost.Populate(); err != nil {
			log.Fatalf("error populate seen posts; %v", err)
		}
		log.Printf("Populated seen posts completed. Posts: %d.\n", populateSeenPost.Posts)
	},
}
