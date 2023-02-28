package run

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	admin "github.com/memocash/index/admin/server"
	"github.com/memocash/index/node"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/broadcast/broadcast_server"
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
		processor := lead.NewProcessor(s.Verbose)
		jlog.Logf("Cluster lead processor starting...\n")
		go func() {
			errorHandler <- jerr.Get("error running cluster lead processor", processor.Run())
		}()
		broadcastServer := broadcast_server.NewServer(config.GetBroadcastRpc().Port, func(ctx context.Context, raw []byte) error {
			txMsg, err := memo.GetMsgFromRaw(raw)
			if err != nil {
				return jerr.Get("error parsing raw tx", err)
			}
			jlog.Logf("Broadcasting transaction: %s\n", txMsg.TxHash())
			if err := processor.BlockNode.Peer.BroadcastTx(ctx, txMsg); err != nil {
				return jerr.Get("error broadcasting tx to connection peer", err)
			}
			return nil
		})
		go func() {
			jlog.Logf("Running broadcast server on port: %d\n", broadcastServer.Port)
			errorHandler <- jerr.Get("fatal error running broadcast server", broadcastServer.Run())
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
