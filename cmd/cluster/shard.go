package cluster

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/cluster/shard"
	"github.com/spf13/cobra"
)

var shardCmd = &cobra.Command{
	Use:   "shard",
	Short: "shard [id]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) == 0 {
			jerr.New("fatal error must specify a shard id").Fatal()
		}
		shardId := jutil.GetIntFromString(args[0])
		if fmt.Sprintf("%d", shardId) != args[0] {
			jerr.New("fatal error invalid shard id").Fatal()
		}
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		l := shard.NewShard(shardId, verbose)
		jerr.Get("fatal error running shard", l.Run()).Fatal()
	},
}
