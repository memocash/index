package cli

import (
	"github.com/jchavannes/jgo/jerr"
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
		messenger := network_client.NewOutputMessenger()
		messenger.Output(args[0])
	},
}
