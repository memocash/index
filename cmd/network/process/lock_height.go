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

var lockHeightCmd = &cobra.Command{
	Use: "lock-height",
	Run: func(c *cobra.Command, args []string) {
		var startHeight int64
		if len(args) >= 1 {
			startHeight = jutil.GetInt64FromString(args[0])
			if startHeight == 0 {
				startHeight = -1
			}
		}
		jlog.Log("Starting lock height processor...")
		shard, _ := c.Flags().GetInt(FlagShard)
		lockHeightStatus := status.NewHeight(status.GetStatusShardName(status.NameLockHeight, shard), startHeight)
		lockHeightSaver := saver.NewLockHeight(false)
		lockHeightProcessor := process.NewBlockShard(shard, lockHeightStatus, lockHeightSaver)
		if err := lockHeightProcessor.Process(); err != nil {
			jerr.Get("fatal error processing lock height", err).Fatal()
		}
	},
}
