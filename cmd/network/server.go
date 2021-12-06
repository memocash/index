package network

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/network_server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Network server",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		port := config.GetServerPort()
		server := network_server.NewServer(verbose, port)
		jlog.Logf("Starting network server on port: %d\n", port)
		err := server.Serve()
		if err != nil {
			jerr.Get("fatal error with network server", err).Fatal()
		}
	},
}
