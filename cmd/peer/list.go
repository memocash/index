package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/client"
	"github.com/memocash/server/db/item"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		peers, err := item.GetPeers(0, nil)
		if err != nil {
			jerr.Get("fatal error getting peers", err).Fatal()
		}
		jlog.Logf("Peers: %d\n", len(peers))
	},
}

var getCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {
		peerGet := client.NewPeerGet()
		if err := peerGet.Get(); err != nil {
			jerr.Get("fatal error getting admin index", err).Fatal()
		}
		jlog.Logf("peerGet.Message: %s\n", peerGet.Message)
	},
}
