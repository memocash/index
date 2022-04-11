package maint

import (
	"github.com/spf13/cobra"
)

const (
	FlagVerbose = "verbose"
)

var maintCommand = &cobra.Command{
	Use: "maint",
}

func GetCommand() *cobra.Command {
	populateDoubleSpendSeenCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	txLostCleanupCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	maintCommand.AddCommand(
		txLostCleanupCmd,
		populateDoubleSpendSeenCmd,
	)
	return maintCommand
}
