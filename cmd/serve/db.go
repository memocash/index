package serve

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	db "github.com/memocash/server/db/server"
	"github.com/memocash/server/ref/config"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "db [shard]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) == 0 {
			jerr.New("fatal error must specify a shard").Fatal()
		}
		shard := jutil.GetIntFromString(args[0])
		shards := config.GetQueueShards()
		if len(shards) < shard {
			jerr.Newf("fatal error shard specified greater than num shards: %d %d", shard, len(shards)).Fatal()
		}
		server := db.NewServer(uint(shards[shard].Port), uint(shard))
		jlog.Logf("Starting queue server shard %d on port %d...\n", server.Shard, server.Port)
		jerr.Get("fatal error running queue server", server.Run()).Fatal()
	},
}
