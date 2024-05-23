package serve

import (
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/cluster/shard"
	"github.com/spf13/cobra"
	"log"
)

var shardCmd = &cobra.Command{
	Use:   "shard",
	Short: "shard [id]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalf("fatal error must specify a shard id")
		}
		shardId := jutil.GetIntFromString(args[0])
		if fmt.Sprintf("%d", shardId) != args[0] {
			log.Fatalf("fatal error invalid shard id")
		}
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		l := shard.NewShard(shardId, verbose)
		log.Fatalf("fatal error running shard; %v", l.Run())
	},
}
