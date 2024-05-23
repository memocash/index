package maint

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var queueProfileCmd = &cobra.Command{
	Use: "queue-profile",
	Run: func(cmd *cobra.Command, args []string) {
		var startHeight int64 = 350000
		if len(args) > 0 {
			startHeight = jutil.GetInt64FromString(args[0])
		}
		const Shard = 0
		heightBlocks, err := chain.GetHeightBlocks(Shard, startHeight, false)
		if err != nil {
			log.Fatalf("fatal error getting height blocks; %v", err)
		}
		var outputs int
		start := time.Now()
		log.Println("starting queue profile...")
		for _, heightBlock := range heightBlocks {
			if db.GetShardIdFromByte(heightBlock.BlockHash[:]) != Shard {
				continue
			}
			blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{
				BlockHash: heightBlock.BlockHash,
				Limit:     client.HugeLimit,
			})
			if err != nil {
				log.Fatalf("fatal error getting block txs; %v", err)
			}
			var uids [][]byte
			for _, blockTx := range blockTxs {
				if db.GetShardIdFromByte(blockTx.TxHash[:]) == Shard {
					uids = append(uids, jutil.ByteReverse(blockTx.TxHash[:]))
				}
			}
			dbClient := client.NewClient(config.GetShardConfig(Shard, config.GetQueueShards()).GetHost())
			if err := dbClient.GetByPrefixes(db.TopicChainTxOutput, uids); err != nil {
				log.Fatalf("fatal error getting db message tx outputs; %v", err)
			}
			log.Printf("%d outputs retrieved for shard 0 (height: %d, txs: %d)\n",
				len(dbClient.Messages), heightBlock.Height, len(blockTxs))
			outputs += len(dbClient.Messages)
		}
		log.Printf("done. (height blocks: %d, outputs: %d, time: %s)\n",
			len(heightBlocks), outputs, time.Now().Sub(start))
	},
}
