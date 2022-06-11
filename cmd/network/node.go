package network

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/peer"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/broadcast/broadcast_server"
	"github.com/memocash/index/ref/config"
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

var broadcasterCmd = &cobra.Command{
	Use:   "broadcaster",
	Short: "Run Network Broadcaster Node",
	Run: func(c *cobra.Command, args []string) {
		connection := peer.NewConnection(nil, nil)
		broadcastServer := broadcast_server.NewServer(config.GetBroadcastRpc().Port, func(ctx context.Context, raw []byte) error {
			txMsg, err := memo.GetMsgFromRaw(raw)
			if err != nil {
				return jerr.Get("error parsing raw tx", err)
			}
			jlog.Logf("Broadcasting transaction: %s\n", txMsg.TxHash())
			if err := connection.BroadcastTx(ctx, txMsg); err != nil {
				return jerr.Get("error broadcasting tx to connection peer", err)
			}
			return nil
		})
		go func() {
			jlog.Logf("Running broadcast server on port: %d\n", broadcastServer.Port)
			err := broadcastServer.Run()
			jerr.Get("fatal error running broadcast server", err).Fatal()
		}()
		if err := connection.Connect(); err != nil {
			jerr.Get("fatal error connecting to peer", err).Fatal()
		}
		jlog.Log("connection ended")
	},
}
