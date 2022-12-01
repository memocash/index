package serve

import "github.com/spf13/cobra"

const FlagVerbose = "verbose"
const FlagDev = "dev"

var serveCmd = &cobra.Command{
	Use: "serve",
}

func GetCommand() *cobra.Command {
	allCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	allCmd.Flags().BoolP(FlagDev, "", false, "Don't connect to bitcoin node")
	leadCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	shardCmd.Flags().BoolP(FlagVerbose, "v", false, "Additional logging")
	serveCmd.AddCommand(
		allCmd,
		dbCmd,
		adminCmd,
		leadCmd,
		shardCmd,
	)
	return serveCmd
}
