package serve

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	admin "github.com/memocash/index/admin/server"
	"github.com/memocash/index/api"
	db "github.com/memocash/index/db/server"
	"github.com/memocash/index/node"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/network_server"
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use: "all",
	Run: func(c *cobra.Command, args []string) {
		var errorHandler = make(chan error)
		// API server
		apiServer := api.NewServer()
		if err := apiServer.Start(); err != nil {
			jerr.Get("fatal error starting api server", err).Fatal()
		}
		jlog.Logf("API (unused REST) server started on port: %d...\n", apiServer.Port)
		go func() {
			errorHandler <- jerr.Get("error running api server", apiServer.Serve())
		}()
		// Admin server
		nodeGroup := node.NewGroup()
		adminServer := admin.NewServer(nodeGroup)
		if err := adminServer.Start(); err != nil {
			jerr.Get("fatal error starting admin server", err).Fatal()
		}
		jlog.Logf("Admin server (including graphql) started on port: %d...\n", adminServer.Port)
		go func() {
			errorHandler <- jerr.Get("error running admin server", adminServer.Serve())
		}()
		// Queue servers
		for i, queueShard := range config.GetQueueShards() {
			queueServer := db.NewServer(uint(queueShard.Port), uint(i))
			if err := queueServer.Start(); err != nil {
				jerr.Getf(err, "fatal error starting db queue server shard %d", queueServer.Shard).Fatal()
			}
			jlog.Logf("Queue server started on port: %d...\n", queueServer.Port)
			go func() {
				errorHandler <- jerr.Getf(queueServer.Serve(), "error running db queue server shard %d",
					queueServer.Shard)
			}()
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
		// Error handler
		jerr.Get("fatal memo server error encountered", <-errorHandler).Fatal()
	},
}
