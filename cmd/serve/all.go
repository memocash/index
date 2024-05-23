package serve

import (
	"github.com/memocash/index/node/obj/run"
	"github.com/spf13/cobra"
	"log"
)

var allCmd = &cobra.Command{
	Use: "all",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		server := run.NewServer(true, verbose)
		log.Fatalf("fatal memo server error encountered (dev); %v", server.Run())
	},
}

var liveCmd = &cobra.Command{
	Use: "live",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		server := run.NewServer(false, verbose)
		log.Fatalf("fatal memo server error encountered (live); %v", server.Run())
	},
}
