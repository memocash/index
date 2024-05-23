package peer

import (
	"github.com/memocash/index/admin/client/peer"
	"github.com/spf13/cobra"
	"log"
)

var disconnectCmd = &cobra.Command{
	Use: "disconnect",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("error must give node id")
		}
		nodeId := args[0]
		peerDisconnect := peer.NewDisconnect()
		if err := peerDisconnect.Disconnect(nodeId); err != nil {
			log.Fatalf("fatal error getting peer disconnect; %v", err)
		}
		log.Printf("peerDisconnect.Message: %s\n", peerDisconnect.Message)
	},
}
