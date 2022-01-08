package maint

import (
	"github.com/spf13/cobra"
)

var maintCommand = &cobra.Command{
	Use: "maint",
}

func GetCommand() *cobra.Command {
	maintCommand.AddCommand(
		txLostCleanupCmd,
	)
	return maintCommand
}
