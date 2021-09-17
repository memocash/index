package cmd

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/api"
	"github.com/memocash/server/cmd/test"
	"github.com/memocash/server/db/queue"
	"github.com/memocash/server/node"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Default")
		var errorHandler = make(chan error)
		go func() {
			err := api.NewServer().Run()
			errorHandler <- jerr.Get("fatal error running api server", err)
		}()
		go func() {
			err := node.NewServer().Run()
			errorHandler <- jerr.Get("fatal error running node server", err)
		}()
		go func() {
			err := queue.NewServer(queue.DefaultShard0Port).Run()
			errorHandler <- jerr.Get("fatal error running db queue server shard 0", err)
		}()
		go func() {
			err := queue.NewServer(queue.DefaultShard1Port).Run()
			errorHandler <- jerr.Get("fatal error running db queue server shard 1", err)
		}()
		jerr.Get("fatal memo server error encountered", <-errorHandler).Fatal()
	},
}

func Execute() error {
	serverCmd.AddCommand(
		test.GetCommand(),
	)
	if err := serverCmd.Execute(); err != nil {
		return jerr.Get("error executing server command", err)
	}
	return nil
}
