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
		blockStatus := status.NewHeight(status.NameBlock, startHeight)
		combinedSaver := saver.NewCombined([]dbi.TxSave{
			saver.NewTx(false),
		})
		blockProcessor := process.NewBlock(blockStatus, combinedSaver)
		blockProcessor.UseRaw = true
		if err := blockProcessor.Process(); err != nil {
			jerr.Get("fatal error processing blocks (new)", err).Fatal()
		}
	},
}
