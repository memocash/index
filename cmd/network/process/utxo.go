package process

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/obj/process"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/obj/status"
	"github.com/spf13/cobra"
)

var utxoCmd = &cobra.Command{
	Use: "utxo",
	Run: func(c *cobra.Command, args []string) {
		var startHeight int64
		if len(args) >= 1 {
			startHeight = jutil.GetInt64FromString(args[0])
			if startHeight == 0 {
				startHeight = -1
			}
		}
		shards, _ := c.Flags().GetIntSlice(FlagShards)
		jlog.Log("Starting utxo processor...")
		utxoStatus := status.NewHeight(status.GetStatusShardName(status.NameUtxo, shards), startHeight)
		utxoSaver := saver.NewUtxo(false)
		utxoProcessor := process.NewBlock(utxoStatus, utxoSaver)
		utxoProcessor.Shards = shards
		if err := utxoProcessor.Process(); err != nil {
			jerr.Get("fatal error processing utxos", err).Fatal()
		}
	},
}
