package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/client"
	"github.com/spf13/cobra"
)

var connectDefaultCmd = &cobra.Command{
	Use: "connect-default",
	Run: func(cmd *cobra.Command, args []string) {
		peerConnect := client.NewPeerConnect()
		if err := peerConnect.Get(); err != nil {
			jerr.Get("fatal error getting peer connect", err).Fatal()
		}
		jlog.Logf("peerConnect.Message: %s\n", peerConnect.Message)
	},
}
