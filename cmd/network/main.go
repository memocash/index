package network

import (
	"github.com/memocash/index/cmd/network/process"
	"github.com/memocash/index/cmd/network/single"
	"github.com/spf13/cobra"
)

const (
	FlagVerbose = "verbose"
)

var networkCommand = &cobra.Command{
	Use: "network",
}

func GetCommand() *cobra.Command {
	nodeCmd.PersistentFlags().Bool(FlagVerbose, false, "verbose")
	mempoolCmd.PersistentFlags().Bool(FlagVerbose, false, "verbose")
	serverCmd.PersistentFlags().Bool(FlagVerbose, false, "verbose")
	networkCommand.AddCommand(
		process.GetCommand(),
		single.GetCommand(),
		nodeCmd,
		mempoolCmd,
		serverCmd,
	)
	return networkCommand
}
