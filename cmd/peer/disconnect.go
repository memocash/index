package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/client/peer"
	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use: "disconnect",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			jerr.Newf("error must give node id").Fatal()
		}
		nodeId := args[0]
		peerDisconnect := peer.NewDisconnect()
		if err := peerDisconnect.Disconnect(nodeId); err != nil {
			jerr.Get("fatal error getting peer disconnect", err).Fatal()
		}
		jlog.Logf("peerDisconnect.Message: %s\n", peerDisconnect.Message)
	},
}
