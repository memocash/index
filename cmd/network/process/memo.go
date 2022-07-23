package process

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/obj/process"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/obj/status"
	"github.com/memocash/index/ref/dbi"
	"github.com/spf13/cobra"
)

var memoCmd = &cobra.Command{
	Use: "memo",
	Run: func(c *cobra.Command, args []string) {
		var startHeight int64
		if len(args) >= 1 {
			startHeight = jutil.GetInt64FromString(args[0])
			if startHeight == 0 {
				startHeight = -1
			}
		}
		jlog.Log("Starting memo processor...")
		shard, _ := c.Flags().GetInt(FlagShard)
		blockStatus := status.NewHeight(status.GetStatusShardName(status.NameMemo, shard), startHeight)
		combinedSaver := saver.NewCombined([]dbi.TxSave{
			saver.NewMemo(false),
		})
		blockProcessor := process.NewBlockShard(shard, blockStatus, combinedSaver)
		if err := blockProcessor.Process(); err != nil {
			jerr.Get("fatal error processing memo blocks", err).Fatal()
		}
	},
}
