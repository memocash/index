package cluster

import "github.com/spf13/cobra"

var clusterCmd = &cobra.Command{
	Use: "cluster",
}

func GetCommand() *cobra.Command {
	clusterCmd.AddCommand(
		leadCmd,
		shardCmd,
	)
	return clusterCmd
}
