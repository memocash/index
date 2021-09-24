package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/client/peer"
	"github.com/spf13/cobra"
)

var connectDefaultCmd = &cobra.Command{
	Use: "connect-default",
	Run: func(cmd *cobra.Command, args []string) {
		peerConnect := peer.NewConnect()
		if err := peerConnect.Get(); err != nil {
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
