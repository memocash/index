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
		jlog.Log("Starting utxo processor...")
		shard, _ := c.Flags().GetInt(FlagShard)
		utxoStatus := status.NewHeight(status.GetStatusShardName(status.NameUtxo, shard), startHeight)
		utxoProcessor := process.NewBlockShard(shard, utxoStatus, saver.NewUtxo(false))
		if err := utxoProcessor.Process(); err != nil {
			jerr.Get("fatal error processing utxos", err).Fatal()
		}
	},
}
