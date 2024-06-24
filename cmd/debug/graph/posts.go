package graph

import (
	"github.com/memocash/index/client/lib/graph"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var postsCmd = &cobra.Command{
	Use: "posts",
	Run: func(c *cobra.Command, args []string) {
		var start time.Time
		if len(args) > 0 {
			var err error
			if start, err = time.Parse("2006-01-02", args[0]); err != nil {
				if start, err = time.Parse(time.RFC3339, args[0]); err != nil {
					log.Fatalf("error parsing start time; %v", err)
				}
			}
		}
		posts, err := graph.GetPosts(start)
		if err != nil {
			log.Fatalf("error getting posts; %v", err)
		}
		for i, post := range posts {
			log.Printf("Post %2d (%s): %s %s %s\n", i, post.Tx.Seen.Format(time.RFC3339Nano),
				post.TxHash, post.Address, post.Text)
		}
	},
}
