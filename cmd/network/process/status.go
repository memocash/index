package process

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/obj/status"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/spf13/cobra"
)

var statusGetCmd = &cobra.Command{
	Use:   "status-get",
	Short: "status-get [name] [shard]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 2 {
			jerr.New("fatal error must specify topic and shard").Fatal()
		}
		topicName := args[0]
		shard := jutil.GetIntFromString(args[1])
		statusShard := status.NewHeight(status.GetStatusShardName(topicName, shard), 0)
		height, err := statusShard.GetHeight()
		if err != nil {
			jerr.Get("fatal error getting height for status shard", err).Fatal()
		}
		jlog.Logf("height: %d, block: %s\n", height.Height, hs.GetTxString(height.Block))
	},
}

var statusSetCmd = &cobra.Command{
	Use:   "status-set",
	Short: "status-set [name] [shard] [height]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 3 {
			jerr.New("fatal error must specify topic, shard, and height").Fatal()
		}
		topicName := args[0]
		shard := jutil.GetIntFromString(args[1])
		height := jutil.GetInt64FromString(args[2])
		statusShard := status.NewHeight(status.GetStatusShardName(topicName, shard), 0)
		heightBlock, err := chain.GetHeightBlockSingle(height)
		if err != nil {
			jerr.Get("fatal error getting height block for setting status", err).Fatal()
		}
		if err := statusShard.SetHeight(status.BlockHeight{
			Height: heightBlock.Height,
			Block:  heightBlock.BlockHash[:],
		}); err != nil {
			jerr.Get("fatal error getting height for status shard", err).Fatal()
		}
		jlog.Logf("set height: %d, block: %s\n", heightBlock.Height, chainhash.Hash(heightBlock.BlockHash))
	},
}
