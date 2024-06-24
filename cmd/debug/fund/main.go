package fund

import "github.com/spf13/cobra"

var fundCmd = &cobra.Command{
	Use: "fund",
}

func GetCommand() *cobra.Command {
	fundCmd.AddCommand(
		addressCmd,
	)
	return fundCmd
}

