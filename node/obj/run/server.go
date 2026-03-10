package run

import (
	"context"
	"fmt"
	"log"

	admin "github.com/memocash/index/admin/server"
	graph "github.com/memocash/index/graph/server"
	"github.com/memocash/index/node"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/broadcast/broadcast_server"
	"github.com/memocash/index/ref/cluster/lead"
	"github.com/memocash/index/ref/cluster/shard"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/network_server"
)

type Server struct {
	Context context.Context
	Dev     bool
	Verbose bool
}

func (s *Server) Run() error {
	var errorHandler = make(chan error)
	// Admin server
	adminServer := admin.NewServer(node.NewGroup(s.Context))
	if err := adminServer.Start(); err != nil {
		return fmt.Errorf("fatal error starting admin server; %w", err)
	}
	log.Printf("AdminServer starting on port: %d\n", adminServer.Port)
	go func() {
		errorHandler <- fmt.Errorf("error running admin server; %w", adminServer.Serve())
	}()
	// GraphQL server
	graphServer := graph.NewServer()
	if err := graphServer.Start(); err != nil {
		return fmt.Errorf("fatal error starting graph server; %w", err)
	}
	log.Printf("GraphQLServer starting on port: %d\n", graphServer.Port)
	go func() {
		errorHandler <- fmt.Errorf("error running GraphQL server; %w", graphServer.Serve())
	}()
	// Cluster shard servers
	for _, shardConfig := range config.GetClusterShards() {
		clusterShard := shard.NewShard(int(shardConfig.Shard), s.Verbose)
		if err := clusterShard.Start(); err != nil {
			return fmt.Errorf("error starting cluster shard %d; %w", shardConfig.Shard, err)
		}
		go func(s *shard.Shard) {
			errorHandler <- fmt.Errorf("error running cluster shard %d; %w", s.Id, clusterShard.Serve())
		}(clusterShard)
	}
	// Network server
	networkServer := network_server.NewServer(false, config.GetServerPort())
	if err := networkServer.Start(); err != nil {
		return fmt.Errorf("fatal error starting network server; %w", err)
	}
	log.Printf("NetworkServer starting on port: %d\n", networkServer.Port)
	go func() {
		errorHandler <- fmt.Errorf("error running network server; %w", networkServer.Serve())
	}()
	if !s.Dev {
		processor := lead.NewProcessor(s.Context, s.Verbose)
		go func() {
			errorHandler <- fmt.Errorf("error running cluster lead processor; %w", processor.Run())
		}()
		broadcastServer := broadcast_server.NewServer(config.GetBroadcastRpc().Port, func(ctx context.Context, raw []byte) error {
			if processor.BlockNode == nil {
				return fmt.Errorf("error BlockNode is nil, cannot setup broadcast server")
			}
			txMsg, err := memo.GetMsgFromRaw(raw)
			if err != nil {
				return fmt.Errorf("error parsing raw tx; %w", err)
			}
			log.Printf("Broadcasting transaction: %s\n", txMsg.TxHash())
			if err := processor.BlockNode.BroadcastTx(ctx, txMsg); err != nil {
				return fmt.Errorf("error broadcasting tx to connection peer; %w", err)
			}
			return nil
		})
		go func() {
			log.Printf("BroadcastServer starting on port: %d\n", broadcastServer.Port)
			errorHandler <- fmt.Errorf("fatal error running broadcast server; %w", broadcastServer.Run())
		}()
	}
	return <-errorHandler
}

func NewServer(ctx context.Context, dev, verbose bool) *Server {
	return &Server{
		Context: ctx,
		Dev:     dev,
		Verbose: verbose,
	}
}
