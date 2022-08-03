package maint

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var checkFollowsCmd = &cobra.Command{
	Use:   "check-follows",
	Short: "check-follows",
	Run: func(c *cobra.Command, args []string) {
		deleteItems, _ := c.Flags().GetBool(FlagDelete)
		checkFollows := maint.NewCheckFollows(deleteItems)
		jlog.Logf("Starting check follows (delete flag: %t)...\n", deleteItems)
		if err := checkFollows.Check(); err != nil {
			jerr.Get("error maint check follows", err).Fatal()
		}
		jlog.Logf("Checked follows: %d, bad: %d\n", checkFollows.Processed, checkFollows.BadFollows)
	},
}
