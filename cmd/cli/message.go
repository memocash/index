package cli

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/network_client"
	"github.com/spf13/cobra"
)

var outputMessageCmd = &cobra.Command{
	Use:   "output-message",
	Short: "output-message [message]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) == 0 {
			jerr.New("fatal error must specify a message").Fatal()
		}
		network_client.SetConfig(config.RpcConfig{
			Host: config.Localhost,
			Port: config.GetServerPort(),
		})
		messenger := network_client.NewOutputMessenger()
		if err := messenger.Output(args[0]); err != nil {
			jerr.Get("error outputting message", err).Fatal()
		}
	},
}
