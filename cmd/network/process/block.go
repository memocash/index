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

var blockCmd = &cobra.Command{
	Use: "block",
	Run: func(c *cobra.Command, args []string) {
		var startHeight int64
		if len(args) >= 1 {
			startHeight = jutil.GetInt64FromString(args[0])
			if startHeight == 0 {
				startHeight = -1
			}
		}
		jlog.Log("Starting block processor...")
		shard, _ := c.Flags().GetInt(FlagShard)
		blockStatus := status.NewHeight(status.GetStatusShardName(status.NameBlock, shard), startHeight)
		txSaver := saver.NewTxShard(false, shard)
		blockProcessor := process.NewBlockRaw(shard, blockStatus, txSaver)
		if err := blockProcessor.Process(); err != nil {
			jerr.Get("fatal error processing blocks (new)", err).Fatal()
		}
	},
}
