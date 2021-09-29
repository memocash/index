package peer

import (
	"github.com/spf13/cobra"
)

var peerCmd = &cobra.Command{
	Use: "peer",
}

func GetCommand() *cobra.Command {
	peerCmd.AddCommand(
		listCmd,
		listFoundPeersCmd,
		listPeerFoundsCmd,
		getCmd,
		connectDefaultCmd,
		connectNextCmd,
		connectCmd,
		listConnectionsCmd,
		historyCmd,
		disconnectCmd,
		loopingEnableCmd,
		loopingDisableCmd,
		foundPeersCmd,
	)
	return peerCmd
}
