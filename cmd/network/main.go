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
	networkCommand.AddCommand(
		process.GetCommand(),
		single.GetCommand(),
		nodeCmd,
		mempoolCmd,
		serverCmd,
	)
	return networkCommand
}
