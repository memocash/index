package network

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/peer"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Run Network Block Node",
	RunE: func(c *cobra.Command, args []string) error {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		connection := peer.NewConnection(saver.CombinedBlockSaver(verbose), saver.CombinedTxSaverNoDS(verbose))
		if err := connection.Connect(); err != nil {
			jerr.Get("fatal error connecting to peer", err).Fatal()
		}
		jlog.Log("connection ended")
		return nil
	},
}

var mempoolCmd = &cobra.Command{
	Use:   "mempool",
	Short: "Run Network Mempool Node",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		connection := peer.NewConnection(nil, saver.CombinedTxSaver(verbose))
		if err := connection.Connect(); err != nil {
			jerr.Get("fatal error connecting to peer", err).Fatal()
		}
		jlog.Log("connection ended")
	},
}
