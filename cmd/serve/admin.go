package serve

import (
	admin "github.com/memocash/index/admin/server"
	"github.com/spf13/cobra"
	"log"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "admin",
	Run: func(c *cobra.Command, args []string) {
		adminServer := admin.NewServer(nil)
		log.Printf("Starting admin server on port %d...\n", adminServer.Port)
		log.Fatalf("fatal error running admin server; %v", adminServer.Run())
	},
}
