package serve

import (
	"github.com/memocash/index/graph/server"
	"github.com/spf13/cobra"
	"log"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "graph",
	Run: func(c *cobra.Command, args []string) {
		graphServer := server.NewServer()
		log.Printf("Starting graph server on port %d...\n", graphServer.Port)
		log.Fatalf("fatal error running graph server; %v", graphServer.Run())
	},
}
