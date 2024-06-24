package debug

import (
	"github.com/memocash/index/cmd/debug/fund"
	"github.com/memocash/index/cmd/debug/graph"
	"github.com/memocash/index/cmd/debug/load"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug commands",
}

func GetCommand() *cobra.Command {
	debugCmd.AddCommand(
		saveTxCmd,
		itemDeleteCmd,
		getTestCommand(),
		fund.GetCommand(),
		graph.GetCommand(),
		load.GetCommand(),
	)
	return debugCmd
}
