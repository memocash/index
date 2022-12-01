package cmd

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/cmd/cli"
	"github.com/memocash/index/cmd/maint"
	"github.com/memocash/index/cmd/network"
	"github.com/memocash/index/cmd/peer"
	"github.com/memocash/index/cmd/serve"
	"github.com/memocash/index/cmd/test"
	"github.com/memocash/index/ref/broadcast/broadcast_client"
	"github.com/memocash/index/ref/config"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
)

var pf interface {
	Stop()
}

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
		broadcast_client.SetConfig(config.GetBroadcastRpc())
		profileExecution, _ := cmd.Flags().GetBool(config.FlagProfile)
		if profileExecution {
			pf = profile.Start()
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if pf != nil {
			pf.Stop()
		}
	},
}

func Execute() error {
	indexCmd.PersistentFlags().String(config.FlagConfig, "", "config file name")
	indexCmd.PersistentFlags().Bool(config.FlagProfile, false, "profile execution")
	indexCmd.AddCommand(
		cli.GetCommand(),
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
