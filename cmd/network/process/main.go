package process

import "github.com/spf13/cobra"

var processCommand = &cobra.Command{
	Use: "process",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func GetCommand() *cobra.Command {
	processCommand.AddCommand(
		nodeCmd,
	)
	return processCommand
}

