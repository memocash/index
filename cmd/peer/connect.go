package peer

import (
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/client/peer"
	"github.com/spf13/cobra"
	"log"
	"net"
)

var connectDefaultCmd = &cobra.Command{
	Use: "connect-default",
	Run: func(cmd *cobra.Command, args []string) {
		peerConnectDefault := peer.NewConnectDefault()
		if err := peerConnectDefault.Get(); err != nil {
			log.Fatalf("fatal error getting peer connect; %v", err)
		}
		log.Printf("peerConnect.Message: %s\n", peerConnectDefault.Message)
	},
}

var connectNextCmd = &cobra.Command{
	Use: "connect-next",
	Run: func(cmd *cobra.Command, args []string) {
		peerConnectNext := peer.NewConnectNext()
		if err := peerConnectNext.Get(); err != nil {
			log.Fatalf("fatal error getting peer connect next; %v", err)
		}
		log.Printf("peerConnectNext.Message: %s\n", peerConnectNext.Message)
	},
}

var connectCmd = &cobra.Command{
	Use: "connect",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatalf("error must give node ip and port")
		}
		ip := net.ParseIP(args[0])
		if ip == nil {
			log.Fatalf("fatal error unable to parse ip")
		}
		port := jutil.GetUInt16FromString(args[1])
		peerConnect := peer.NewConnect()
		if err := peerConnect.Connect(ip, port); err != nil {
			log.Fatalf("fatal error getting peer connect; %v", err)
		}
		log.Printf("peerConnect.Message: %s\n", peerConnect.Message)
	},
}

var listConnectionsCmd = &cobra.Command{
	Use: "list-connections",
	Run: func(cmd *cobra.Command, args []string) {
		listConnections := peer.NewListConnections()
		if err := listConnections.List(); err != nil {
			log.Fatalf("fatal error getting peer connection list; %v", err)
		}
		log.Printf("listConnections.Connections:\n%s\n", listConnections.Connections)
	},
}

var historyCmd = &cobra.Command{
	Use: "history",
	Run: func(cmd *cobra.Command, args []string) {
		history := peer.NewHistory()
		if err := history.Get(); err != nil {
			log.Fatalf("fatal error getting peer history; %v", err)
		}
		log.Printf("history.Connections (%d):\n", len(history.Connections))
		for i := 0; i < len(history.Connections) && i < 10; i++ {
			conn := history.Connections[i]
			fmt.Printf("Peer connection: %s:%d - %s - %d\n", conn.Ip, conn.Port,
				conn.Time.Format("2006-01-02 15:04:05"), conn.Status)
		}
	},
}
