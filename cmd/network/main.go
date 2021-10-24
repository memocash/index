package network

import (
	"github.com/memocash/server/cmd/network/process"
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
		nodeCmd,
		mempoolCmd,
	)
	return networkCommand
}
