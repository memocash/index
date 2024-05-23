package serve

import "github.com/spf13/cobra"

const FlagVerbose = "verbose"

var serveCmd = &cobra.Command{
	Use: "serve",
}

func GetCommand() *cobra.Command {
	allCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	liveCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	leadCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	networkCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	shardCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	serveCmd.AddCommand(
		allCmd,
		liveCmd,
		dbCmd,
		adminCmd,
		graphCmd,
		broadcasterCmd,
		networkCmd,
		leadCmd,
		shardCmd,
	)
	return serveCmd
}
