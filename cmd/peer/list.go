package peer

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/client/peer"
	"github.com/memocash/index/db/item"
	"github.com/spf13/cobra"
	"log"
	"net"
)

var listCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		var shard uint32
		if len(args) > 0 {
			shard = jutil.GetUInt32FromString(args[0])
		}
		peers, err := item.GetPeers(shard, nil)
		if err != nil {
			log.Fatalf("fatal error getting peers; %v", err)
		}
		log.Printf("Peers: %d\n", len(peers))
		for i := 0; i < len(peers) && i < 10; i++ {
			log.Printf("Peer: %s:%d - %d\n", net.IP(peers[i].Ip), peers[i].Port, peers[i].Services)
		}
	},
}

var listFoundPeersCmd = &cobra.Command{
	Use: "list-found-peers",
	Run: func(cmd *cobra.Command, args []string) {
		var shard uint32
		if len(args) > 0 {
			shard = jutil.GetUInt32FromString(args[0])
		}
		foundPeers, err := item.GetFoundPeers(shard, nil, nil, 0)
		if err != nil {
			log.Fatalf("fatal error getting found peers; %v", err)
		}
		log.Printf("Found peers: %d\n", len(foundPeers))
		for i := 0; i < len(foundPeers) && i < 10; i++ {
			log.Printf("Found peer: %s:%d - %s:%d\n", net.IP(foundPeers[i].Ip), foundPeers[i].Port,
				net.IP(foundPeers[i].FoundIp), foundPeers[i].FoundPort)
		}
	},
}

var listPeerFoundsCmd = &cobra.Command{
	Use: "list-peer-founds",
	Run: func(cmd *cobra.Command, args []string) {
		var shard uint32
		if len(args) > 0 {
			shard = jutil.GetUInt32FromString(args[0])
		}
		foundPeers, err := item.GetPeerFounds(shard, nil)
		if err != nil {
			log.Fatalf("fatal error getting peer founds; %v", err)
		}
		log.Printf("Peer founds: %d\n", len(foundPeers))
		for i := 0; i < len(foundPeers) && i < 10; i++ {
			log.Printf("Peer founds: %s:%d - %s:%d\n", net.IP(foundPeers[i].Ip), foundPeers[i].Port,
				net.IP(foundPeers[i].FinderIp), foundPeers[i].FinderPort)
		}
	},
}

var listAttemptsCmd = &cobra.Command{
	Use: "list-attempts",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("error must give peer address")
		}
		host, portString, err := net.SplitHostPort(args[0])
		if err != nil {
			log.Fatalf("error splitting input host port; %v", err)
		}
		ip := net.ParseIP(host)
		if ip == nil {
			log.Fatalf("error parsing host ip")
		}
		port := jutil.GetUInt16FromString(portString)
		lastPeerConnection, err := item.GetPeerConnectionLast(ip, port)
		if err != nil {
			log.Fatalf("fatal error last peer connection; %v", err)
		}
		log.Printf("lastPeerConnection: %s:%d - %s %s\n", net.IP(lastPeerConnection.Ip), lastPeerConnection.Port,
			lastPeerConnection.Time.Format("2006-01-02 15:04:05"), lastPeerConnection.Status)
	},
}

var getCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {
		peerGet := peer.NewGet()
		if err := peerGet.Get(); err != nil {
			log.Fatalf("fatal error getting peer get; %v", err)
		}
		log.Printf("peerGet.Message: %s\n", peerGet.Message)
	},
}
