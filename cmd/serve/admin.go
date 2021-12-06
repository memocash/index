package serve

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	admin "github.com/memocash/index/admin/server"
	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "admin",
	Run: func(c *cobra.Command, args []string) {
		adminServer := admin.NewServer(nil)
		jlog.Logf("Starting admin server on port %d...\n", adminServer.Port)
		jerr.Get("fatal error running admin server", adminServer.Run()).Fatal()
	},
}
