package cmd

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/cmd/maint"
	"github.com/memocash/index/cmd/network"
	"github.com/memocash/index/cmd/peer"
	"github.com/memocash/index/cmd/serve"
	"github.com/memocash/index/cmd/test"
	"github.com/memocash/index/ref/config"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Run Index Server",
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
	indexCmd.PersistentFlags().String(config.FlagConfig, "", "config file name")
	indexCmd.AddCommand(
		test.GetCommand(),
		peer.GetCommand(),
		network.GetCommand(),
		serve.GetCommand(),
		maint.GetCommand(),
	)
	if err := indexCmd.Execute(); err != nil {
		return jerr.Get("error executing server command", err)
	}
	return nil
}
