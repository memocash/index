package serve

import (
	"context"
	"fmt"
	"github.com/memocash/index/node/peer"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/broadcast/broadcast_server"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/network_server"
	"github.com/spf13/cobra"
	"log"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Run Network server",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		port := config.GetServerPort()
		server := network_server.NewServer(verbose, port)
		log.Printf("Starting network server on port: %d\n", port)
		if err := server.Run(); err != nil {
			log.Fatalf("fatal error with network server; %v", err)
		}
	},
}

var broadcasterCmd = &cobra.Command{
	Use:   "broadcaster",
	Short: "Run Network Broadcaster Node",
	Run: func(c *cobra.Command, args []string) {
		connection := peer.NewConnection(nil, nil)
		broadcastServer := broadcast_server.NewServer(config.GetBroadcastRpc().Port, func(ctx context.Context, raw []byte) error {
			txMsg, err := memo.GetMsgFromRaw(raw)
			if err != nil {
				return fmt.Errorf("error parsing raw tx; %w", err)
			}
			log.Printf("Broadcasting transaction: %s\n", txMsg.TxHash())
			if err := connection.BroadcastTx(ctx, txMsg); err != nil {
				return fmt.Errorf("error broadcasting tx to connection peer; %w", err)
			}
			return nil
		})
		go func() {
			log.Printf("Running broadcast server on port: %d\n", broadcastServer.Port)
			err := broadcastServer.Run()
			log.Fatalf("fatal error running broadcast server; %v", err)
		}()
		if err := connection.Connect(); err != nil {
			log.Fatalf("fatal error connecting to peer; %v", err)
		}
		log.Println("connection ended")
	},
}
