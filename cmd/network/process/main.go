package process

import "github.com/spf13/cobra"

var processCommand = &cobra.Command{
	Use: "process",
}

func GetCommand() *cobra.Command {
	processCommand.AddCommand(
		blockCmd,
		doubleSpendCmd,
		lockHeightCmd,
	)
	return processCommand
}
