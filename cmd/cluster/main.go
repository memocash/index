package cluster

import "github.com/spf13/cobra"

const FlagVerbose = "verbose"

var clusterCmd = &cobra.Command{
	Use: "cluster",
}

func GetCommand() *cobra.Command {
	leadCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	shardCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	clusterCmd.AddCommand(
		leadCmd,
		shardCmd,
	)
	return clusterCmd
}
