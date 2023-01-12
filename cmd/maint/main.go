package maint

import (
	"github.com/spf13/cobra"
)

const (
	FlagVerbose = "verbose"
	FlagDelete  = "delete"
)

var maintCommand = &cobra.Command{
	Use: "maint",
}

func GetCommand() *cobra.Command {
	populateDoubleSpendSeenCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	txLostCleanupCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	checkLockUtxoCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	checkFollowsCmd.Flags().BoolP(FlagDelete, "", false, "Delete items")
	maintCommand.AddCommand(
		txLostCleanupCmd,
		populateDoubleSpendSeenCmd,
		populateHeightBlockShardCmd,
		queueProfileCmd,
		checkLockUtxoCmd,
		checkFollowsCmd,
		populateP2shCmd,
	)
	return maintCommand
}
