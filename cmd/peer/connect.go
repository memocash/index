package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/admin/client/peer"
	"github.com/spf13/cobra"
	"net"
)

var connectDefaultCmd = &cobra.Command{
	Use: "connect-default",
	Run: func(cmd *cobra.Command, args []string) {
		peerConnectDefault := peer.NewConnectDefault()
		if err := peerConnectDefault.Get(); err != nil {
			jerr.Get("fatal error getting peer connect", err).Fatal()
		}
		jlog.Logf("peerConnect.Message: %s\n", peerConnectDefault.Message)
	},
}

var connectNextCmd = &cobra.Command{
	Use: "connect-next",
	Run: func(cmd *cobra.Command, args []string) {
		peerConnectNext := peer.NewConnectNext()
		if err := peerConnectNext.Get(); err != nil {
			jerr.Get("fatal error getting peer connect next", err).Fatal()
		}
		jlog.Logf("peerConnectNext.Message: %s\n", peerConnectNext.Message)
	},
}

var connectCmd = &cobra.Command{
	Use: "connect",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			jerr.Newf("error must give node ip and port").Fatal()
		}
		ip := net.ParseIP(args[0])
		if ip == nil {
			jerr.Newf("fatal error unable to parse ip").Fatal()
		}
		port := jutil.GetUInt16FromString(args[1])
		peerConnect := peer.NewConnect()
		if err := peerConnect.Connect(ip, port); err != nil {
			jerr.Get("fatal error getting peer connect", err).Fatal()
		}
		jlog.Logf("peerConnect.Message: %s\n", peerConnect.Message)
	},
}

var listConnectionsCmd = &cobra.Command{
	Use: "list-connections",
	Run: func(cmd *cobra.Command, args []string) {
		listConnections := peer.NewListConnections()
		if err := listConnections.List(); err != nil {
			jerr.Get("fatal error getting peer connection list", err).Fatal()
		}
		jlog.Logf("listConnections.Connections:\n%s\n", listConnections.Connections)
	},
}

var historyCmd = &cobra.Command{
	Use: "history",
	Run: func(cmd *cobra.Command, args []string) {
		history := peer.NewHistory()
		if err := history.Get(); err != nil {
			jerr.Get("fatal error getting peer history", err).Fatal()
		}
		jlog.Logf("history.Connections:\n%s\n", history.Connections)
	},
}
