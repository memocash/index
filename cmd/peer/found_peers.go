package peer

import (
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/client/peer"
	"github.com/spf13/cobra"
	"log"
	"net"
)

var foundPeersCmd = &cobra.Command{
	Use: "found-peers",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("error must give peer address")
		}
		host, portString, err := net.SplitHostPort(args[0])
		if err != nil {
			fmt.Errorf("error splitting input host port; %w", err)
		}
		ip := net.ParseIP(host)
		if ip == nil {
			log.Println("error parsing host ip")
		}
		port := jutil.GetUInt16FromString(portString)
		foundPeers := peer.NewFoundPeers()
		if err := foundPeers.Get(ip, port); err != nil {
			log.Fatalf("fatal error getting found peers; %v", err)
		}
		log.Printf("foundPeers.FoundPeers: %d\n", len(foundPeers.FoundPeers))
		for i := 0; i < len(foundPeers.FoundPeers) && i < 10; i++ {
			peer := foundPeers.FoundPeers[i]
			log.Printf("peer: %s:%d\n", net.IP(peer.FoundIp), peer.FoundPort)
		}
	},
}
