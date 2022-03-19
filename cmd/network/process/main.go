package process

import "github.com/spf13/cobra"

const FlagShards = "shards"

var processCommand = &cobra.Command{
	Use: "process",
}

func GetCommand() *cobra.Command {
	doubleSpendCmd.PersistentFlags().IntSlice(FlagShards, nil, "--shards 1,2,3")
	processCommand.AddCommand(
		blockCmd,
		doubleSpendCmd,
		lockHeightCmd,
	)
	return processCommand
}
