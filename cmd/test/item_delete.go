package test

import (
	"encoding/hex"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"github.com/spf13/cobra"
	"log"
	"strconv"
)

var itemDeleteCmd = &cobra.Command{
	Use:   "item-delete",
	Short: "item-delete [shard] [topic] [uid]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 3 {
			log.Fatalf("not enough arguments, must specify shard, topic and UID")
		}
		shard, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("error parsing shard; %v", err)
		}
		topic := args[1]
		uid, err := hex.DecodeString(args[2])
		if err != nil {
			log.Fatalf("error decoding uid; %v", err)
		}
		queueShards := config.GetQueueShards()
		shardConfig := config.GetShardConfig(uint32(shard), queueShards)
		db := client.NewClient(shardConfig.GetHost())
		if err := db.DeleteMessages(topic, [][]byte{uid}); err != nil {
			log.Fatalf("error deleting shard topic item: %d %s; %v", shard, topic, err)
		}
	},
}
