package cli

import "github.com/spf13/cobra"

var cliCmd = &cobra.Command{
	Use: "cli",
}

func GetCommand() *cobra.Command {
	cliCmd.AddCommand(
		outputMessageCmd,
	)
	return cliCmd
}
