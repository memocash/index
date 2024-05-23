package cmd

import (
	"fmt"
	"github.com/memocash/index/cmd/maint"
	"github.com/memocash/index/cmd/peer"
	"github.com/memocash/index/cmd/serve"
	"github.com/memocash/index/cmd/test"
	"github.com/memocash/index/db/store"
	"github.com/memocash/index/ref/broadcast/broadcast_client"
	"github.com/memocash/index/ref/config"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"syscall"
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
		config.SetLogger()
		if err := config.Init(cmd); err != nil {
			log.Fatalf("fatal error initializing config; %v", err)
		}
		broadcast_client.SetConfig(config.GetBroadcastRpc())
		profileExecution, _ := cmd.Flags().GetBool(config.FlagProfile)
		if profileExecution {
			pf = profile.Start()
		}
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGTERM)
		go func() {
			<-sigc
			store.CloseAll()
			log.Printf("Index server caught SIGTERM, stopping...\n")
			os.Exit(0)
		}()
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
		test.GetCommand(),
		peer.GetCommand(),
		serve.GetCommand(),
		maint.GetCommand(),
	)
	if err := indexCmd.Execute(); err != nil {
		return fmt.Errorf("error executing server command; %w", err)
	}
	return nil
}
