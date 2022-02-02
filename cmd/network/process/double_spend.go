package process

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/obj/process"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/obj/status"
	"github.com/memocash/index/ref/config"
	"github.com/spf13/cobra"
)

var doubleSpendCmd = &cobra.Command{
	Use: "double-spend",
	Run: func(c *cobra.Command, args []string) {
		var startHeight int64
		if len(args) >= 1 {
			startHeight = jutil.GetInt64FromString(args[0])
			if startHeight == 0 {
				startHeight = -1
			}
		}
		jlog.Log("Starting double spend processor...")
		doubleSpendStatus := status.NewHeight(status.NameDoubleSpend, startHeight)
		doubleSpendSaver := saver.NewDoubleSpend(false)
		doubleSpendProcessor := process.NewBlock(doubleSpendStatus, doubleSpendSaver)
		doubleSpendProcessor.Delay = int(config.GetBlocksToConfirm())
		if err := doubleSpendProcessor.Process(); err != nil {
			jerr.Get("fatal error processing double spends", err).Fatal()
		}
	},
}
