package serve

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/node/obj/run"
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use: "all",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		server := run.NewServer(true, verbose)
		jerr.Get("fatal memo server error encountered (dev)", server.Run()).Fatal()
	},
}

var liveCmd = &cobra.Command{
	Use: "live",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		server := run.NewServer(false, verbose)
		jerr.Get("fatal memo server error encountered (live)", server.Run()).Fatal()
	},
}
