package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/admin/client/peer"
	"github.com/spf13/cobra"
	"net"
)

var foundPeersCmd = &cobra.Command{
	Use: "found-peers",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			jerr.Newf("error must give peer address").Fatal()
		}
		host, portString, err := net.SplitHostPort(args[0])
		if err != nil {
			jerr.Get("error splitting input host port", err)
		}
		ip := net.ParseIP(host)
		if ip == nil {
			jerr.New("error parsing host ip")
		}
		port := jutil.GetUInt16FromString(portString)
		foundPeers := peer.NewFoundPeers()
		if err := foundPeers.Get(ip, port); err != nil {
			jerr.Get("fatal error getting found peers", err).Fatal()
		}
		jlog.Logf("foundPeers.FoundPeers: %d\n", len(foundPeers.FoundPeers))
		for i := 0; i < len(foundPeers.FoundPeers) && i < 10; i++ {
			peer := foundPeers.FoundPeers[i]
			jlog.Logf("peer: %s:%d\n", net.IP(peer.FoundIp), peer.FoundPort)
		}
	},
}
