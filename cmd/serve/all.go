package serve

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	admin "github.com/memocash/index/admin/server"
	"github.com/memocash/index/node"
	"github.com/memocash/index/ref/cluster/lead"
	"github.com/memocash/index/ref/cluster/shard"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/network_server"
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use: "all",
	Run: func(c *cobra.Command, args []string) {
		var errorHandler = make(chan error)
		// Admin server
		adminServer := admin.NewServer(node.NewGroup())
		if err := adminServer.Start(); err != nil {
			jerr.Get("fatal error starting admin server", err).Fatal()
		}
		jlog.Logf("Admin server (including graphql) started on port: %d...\n", adminServer.Port)
		go func() {
			errorHandler <- jerr.Get("error running admin server", adminServer.Serve())
		}()
		// Cluster shard servers
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		for _, shardConfig := range config.GetClusterShards() {
			clusterShard := shard.NewShard(int(shardConfig.Shard), verbose)
			if err := clusterShard.Start(); err != nil {
				jerr.Getf(err, "fatal error starting cluster shard %d", shardConfig.Shard).Fatal()
			}
			jlog.Logf("Cluster shard started on port: %d...\n", shardConfig.Port)
			go func(s *shard.Shard) {
				errorHandler <- jerr.Getf(clusterShard.Serve(), "error running cluster shard %d", s.Id)
			}(clusterShard)
		}
		// Network server
		networkServer := network_server.NewServer(false, config.GetServerPort())
		if err := networkServer.Start(); err != nil {
			jerr.Get("fatal error starting network server", err).Fatal()
		}
		jlog.Logf("Starting network server on port: %d\n", networkServer.Port)
		go func() {
			errorHandler <- jerr.Get("error running network server", networkServer.Serve())
		}()
		devMode, _ := c.Flags().GetBool(FlagDev)
		if !devMode {
			clusterLead := lead.NewLead(verbose)
			if err := clusterLead.Start(); err != nil {
				jerr.Get("fatal error starting cluster lead", err).Fatal()
			}
			jlog.Logf("Cluster lead started on port: %d...\n", clusterLead.Port)
			go func() {
				errorHandler <- jerr.Get("error running cluster lead", clusterLead.Serve())
			}()
		}
		// Error handler
		jerr.Get("fatal memo server error encountered", <-errorHandler).Fatal()
	},
}
