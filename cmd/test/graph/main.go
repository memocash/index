package graph

import "github.com/spf13/cobra"

var graphCmd = &cobra.Command{
	Use: "graph",
}

func GetCommand() *cobra.Command {
	graphCmd.AddCommand(
		txCmd,
		postsCmd,
	)
	return graphCmd
}

