package serve

import "github.com/spf13/cobra"

var serveCmd = &cobra.Command{
	Use: "serve",
}

func GetCommand() *cobra.Command {
	serveCmd.AddCommand(
		allCmd,
		dbCmd,
		adminCmd,
	)
	return serveCmd
}
