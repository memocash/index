package single

import "github.com/spf13/cobra"

var singleCommand = &cobra.Command{
	Use: "single",
}

func GetCommand() *cobra.Command {
	singleCommand.AddCommand(
		doubleSpendBlockCmd,
	)
	return singleCommand
}
