package cmd

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/cmd/network"
	"github.com/memocash/server/cmd/peer"
	"github.com/memocash/server/cmd/serve"
	"github.com/memocash/server/cmd/test"
	"github.com/memocash/server/ref/config"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Server",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := config.Init(cmd); err != nil {
			jerr.Get("fatal error initializing config", err).Fatal()
		}
	},
}

func Execute() error {
	serverCmd.AddCommand(
		test.GetCommand(),
		peer.GetCommand(),
		network.GetCommand(),
		serve.GetCommand(),
	)
	if err := serverCmd.Execute(); err != nil {
		return jerr.Get("error executing server command", err)
	}
	return nil
}
