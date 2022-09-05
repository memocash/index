package maint

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"github.com/spf13/cobra"
)

var populateHeightBlockShardCmd = &cobra.Command{
	Use: "populate-height-block-shard",
	Run: func(c *cobra.Command, args []string) {
		shardConfigs := config.GetQueueShards()
		var maxHeight int64
		jlog.Log("starting populate height block shard...")
		for {
			heightBlocks, err := item.GetHeightBlocksAllLimit(maxHeight, false, client.HugeLimit, false)
			if err != nil {
				jerr.Get("fatal error getting height blocks all for populate shards", err).Fatal()
			}
			var newHeightBlockShards []db.Object
			for _, heightBlock := range heightBlocks {
				if heightBlock.Height > maxHeight {
					maxHeight = heightBlock.Height
				}
				for _, shardConfig := range shardConfigs {
					newHeightBlockShards = append(newHeightBlockShards, &item.HeightBlockShard{
						Height:    heightBlock.Height,
						BlockHash: heightBlock.BlockHash,
						Shard:     uint(shardConfig.Shard),
					})
				}
			}
			if err := db.Save(newHeightBlockShards); err != nil {
				jerr.Get("fatal error saving new block height shards", err)
			}
			jlog.Logf("Saved %d height block shards, max height: %d\n", len(newHeightBlockShards), maxHeight)
			if len(heightBlocks) < 0.8*client.HugeLimit {
				break
			}
		}
		jlog.Log("done.")
	},
}
