package run

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	admin "github.com/memocash/index/admin/server"
	"github.com/memocash/index/node"
	"github.com/memocash/index/ref/cluster/lead"
	"github.com/memocash/index/ref/cluster/shard"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/network_server"
)

type Server struct {
	Dev     bool
	Verbose bool
}

func (s *Server) Run() error {
	var errorHandler = make(chan error)
	// Admin server
	adminServer := admin.NewServer(node.NewGroup())
	if err := adminServer.Start(); err != nil {
		return jerr.Get("fatal error starting admin server", err)
	}
	jlog.Logf("Admin server (including graphql) started on port: %d...\n", adminServer.Port)
	go func() {
		errorHandler <- jerr.Get("error running admin server", adminServer.Serve())
	}()
	// Cluster shard servers
	for _, shardConfig := range config.GetClusterShards() {
		clusterShard := shard.NewShard(int(shardConfig.Shard), s.Verbose)
		if err := clusterShard.Start(); err != nil {
			return jerr.Getf(err, "fatal error starting cluster shard %d", shardConfig.Shard)
		}
		jlog.Logf("Cluster shard started on port: %d...\n", shardConfig.Port)
		go func(s *shard.Shard) {
			errorHandler <- jerr.Getf(clusterShard.Serve(), "error running cluster shard %d", s.Id)
		}(clusterShard)
	}
	// Network server
	networkServer := network_server.NewServer(false, config.GetServerPort())
	if err := networkServer.Start(); err != nil {
		return jerr.Get("fatal error starting network server", err)
	}
	jlog.Logf("Starting network server on port: %d\n", networkServer.Port)
	go func() {
		errorHandler <- jerr.Get("error running network server", networkServer.Serve())
	}()
	if !s.Dev {
		clusterLead := lead.NewLead(s.Verbose)
		if err := clusterLead.Start(); err != nil {
			return jerr.Get("fatal error starting cluster lead", err)
		}
		jlog.Logf("Cluster lead started on port: %d...\n", clusterLead.Port)
		go func() {
			errorHandler <- jerr.Get("error running cluster lead", clusterLead.Serve())
		}()
	}
	return <-errorHandler
}

func NewServer(dev, verbose bool) *Server {
	return &Server{
		Dev:     dev,
		Verbose: verbose,
	}
}
