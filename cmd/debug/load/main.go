package load

import "github.com/spf13/cobra"

var loadCmd = &cobra.Command{
	Use: "load",
}

func GetCommand() *cobra.Command {
	loadCmd.AddCommand(
		historyCmd,
	)
	return loadCmd
}

