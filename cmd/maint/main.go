package maint

import (
	"github.com/spf13/cobra"
)

const (
	FlagVerbose = "verbose"
	FlagDelete  = "delete"
	FlagRestart = "restart"
)

var maintCommand = &cobra.Command{
	Use: "maint",
}

func GetCommand() *cobra.Command {
	populateDoubleSpendSeenCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	txLostCleanupCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	checkLockUtxoCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	checkFollowsCmd.Flags().BoolP(FlagDelete, "", false, "Delete items")
	populateP2shDirectCmd.Flags().BoolP(FlagRestart, "", false, "Restart from beginning")
	populateAddrOutputsCmd.Flags().BoolP(FlagRestart, "", false, "Restart from beginning")
	maintCommand.AddCommand(
		txLostCleanupCmd,
		populateDoubleSpendSeenCmd,
		populateHeightBlockShardCmd,
		queueProfileCmd,
		checkLockUtxoCmd,
		checkFollowsCmd,
		populateP2shCmd,
		populateP2shDirectCmd,
		populateAddrOutputsCmd,
	)
	return maintCommand
}
