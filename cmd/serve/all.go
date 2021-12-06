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
		nodeGroup := node.NewGroup()
		apiServer := api.NewServer()
		adminServer := admin.NewServer(nodeGroup)
		var queueServers []*db.Server
		for i, queueShard := range config.GetQueueShards() {
			queueServers = append(queueServers, db.NewServer(uint(queueShard.Port), uint(i)))
		}
		networkServer := network_server.NewServer(false, config.GetServerPort())
		go func() {
			err := apiServer.Run()
			errorHandler <- jerr.Get("error running api server", err)
		}()
		go func() {
			err := adminServer.Run()
			errorHandler <- jerr.Get("error running admin server", err)
		}()
		for i := range queueServers {
			queueServer := queueServers[i]
			go func() {
				err := queueServer.Run()
				errorHandler <- jerr.Getf(err, "error running db queue server shard %d", queueServer.Shard)
			}()
		}
		go func() {
			err := networkServer.Serve()
			errorHandler <- jerr.Get("error running network server", err)
		}()
		jlog.Logf("API (unused REST) server started on port: %d...\n", apiServer.Port)
		jlog.Logf("Admin server (including graphql) started on port: %d...\n", adminServer.Port)
		for i, queueServer := range queueServers {
			jlog.Logf("Queue server %d started on port: %d...\n", i, queueServer.Port)
		}
		jlog.Logf("Starting network server on port: %d\n", networkServer.Port)
		jerr.Get("fatal memo server error encountered", <-errorHandler).Fatal()
	},
}
