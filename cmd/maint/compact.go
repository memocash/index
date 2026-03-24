package maint

import (
	"log"

	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"github.com/spf13/cobra"
)

var compactCmd = &cobra.Command{
	Use:   "compact",
	Short: "Force LevelDB compaction on all shard topics",
	Run: func(c *cobra.Command, args []string) {
		for _, shardConfig := range config.GetQueueShards() {
			host := shardConfig.GetHost()
			log.Printf("Sending compact request to shard %d (%s)...\n", shardConfig.Shard, host)
			dbClient := client.NewClient(host)
			if err := dbClient.CompactAll(); err != nil {
				log.Fatalf("fatal error sending compact to shard %d; %v", shardConfig.Shard, err)
			}
		}
		log.Println("Compaction started on all shards (monitor shard logs for progress)")
	},
}
