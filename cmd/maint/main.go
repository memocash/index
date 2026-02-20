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
	checkFollowsCmd.Flags().BoolP(FlagDelete, "", false, "Delete items")
	populateP2shDirectCmd.Flags().BoolP(FlagRestart, "", false, "Restart from beginning")
	populateAddrOutputsCmd.Flags().BoolP(FlagRestart, "", false, "Restart from beginning")
	populateAddrInputsCmd.Flags().BoolP(FlagRestart, "", false, "Restart from beginning")
	backfillCmd.Flags().Int64(FlagStart, 0, "Start height (required)")
	backfillCmd.Flags().Int64(FlagEnd, 0, "End height (required)")
	backfillCmd.MarkFlagRequired(FlagStart)
	backfillCmd.MarkFlagRequired(FlagEnd)
	maintCommand.AddCommand(
		queueProfileCmd,
		checkFollowsCmd,
		populateP2shCmd,
		populateP2shDirectCmd,
		populateAddrOutputsCmd,
		populateAddrInputsCmd,
		populateSeenPostsCmd,
		doubleSpendCmd,
		randomDoubleSpendCmd,
		scanHeadersCmd,
		backfillCmd,
	)
	return maintCommand
}
