package process

import (
	"github.com/memocash/index/node/obj/status"
	"github.com/spf13/cobra"
)

const (
	FlagShard = "shard"
	FlagDelay = "delay"
)

var processCommand = &cobra.Command{
	Use: "process",
}

func GetCommand() *cobra.Command {
	blockCmd.PersistentFlags().Int(FlagShard, status.NoShard, "--shard 1")
	doubleSpendCmd.PersistentFlags().Int(FlagShard, status.NoShard, "--shard 1")
	doubleSpendCmd.PersistentFlags().Int(FlagDelay, 0, "delay")
	utxoCmd.PersistentFlags().Int(FlagShard, status.NoShard, "--shard 1")
	lockHeightCmd.PersistentFlags().Int(FlagShard, status.NoShard, "--shard 1")
	processCommand.AddCommand(
		blockCmd,
		doubleSpendCmd,
		lockHeightCmd,
		utxoCmd,
		statusGetCmd,
		statusSetCmd,
	)
	return processCommand
}
