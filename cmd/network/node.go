package network

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/peer"
	"github.com/memocash/index/ref/dbi"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Run Network Block Node",
	RunE: func(c *cobra.Command, args []string) error {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		connection := peer.NewConnection(saver.NewCombined([]dbi.TxSave{
			saver.NewTxRaw(verbose),
		}), saver.BlockSaver(verbose))
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
		connection := peer.NewConnection(saver.NewCombined([]dbi.TxSave{
			saver.NewTxRaw(verbose),
			saver.NewTx(verbose),
			saver.NewUtxo(verbose),
			saver.NewLockHeight(verbose),
			saver.NewDoubleSpend(verbose),
		}), nil)
		if err := connection.Connect(); err != nil {
			jerr.Get("fatal error connecting to peer", err).Fatal()
		}
		jlog.Log("connection ended")
	},
}
