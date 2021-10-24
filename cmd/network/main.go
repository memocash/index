package network

import "github.com/spf13/cobra"

const (
	FlagVerbose = "verbose"
)

var networkCommand = &cobra.Command{
	Use: "network",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func GetCommand() *cobra.Command {
	networkCommand.AddCommand(
		nodeCmd,
		mempoolCmd,
	)
	return networkCommand
}
