package process

import "github.com/spf13/cobra"

const (
	FlagShards = "shards"
	FlagDelay  = "delay"
)

var processCommand = &cobra.Command{
	Use: "process",
}

func GetCommand() *cobra.Command {
	doubleSpendCmd.PersistentFlags().IntSlice(FlagShards, nil, "--shards 1,2,3")
	doubleSpendCmd.PersistentFlags().Int(FlagDelay, 0, "delay")
	utxoCmd.PersistentFlags().IntSlice(FlagShards, nil, "--shards 1,2,3")
	processCommand.AddCommand(
		blockCmd,
		doubleSpendCmd,
		lockHeightCmd,
		utxoCmd,
	)
	return processCommand
}
