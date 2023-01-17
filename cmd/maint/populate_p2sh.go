package maint

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var populateP2shCmd = &cobra.Command{
	Use:   "populate-p2sh",
	Short: "populate-p2sh",
	Run: func(c *cobra.Command, args []string) {
		var startHeight int64
		if len(args) > 0 {
			startHeight = jutil.GetInt64FromString(args[0])
		}
		populateP2sh := maint.NewPopulateP2sh()
		jlog.Logf("Starting populate p2sh...\n")
		if err := populateP2sh.Populate(startHeight); err != nil {
			jerr.Get("error populate p2sh", err).Fatal()
		}
		jlog.Logf("Populated p2sh, blocks processed: %d\n", populateP2sh.BlocksProcessed)
	},
}

var populateP2shDirectCmd = &cobra.Command{
	Use:   "populate-p2sh-direct",
	Short: "populate-p2sh-direct",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populateP2sh := maint.NewPopulateP2shDirect()
		jlog.Logf("Starting populate p2sh...\n")
		if err := populateP2sh.Populate(restart); err != nil {
			jerr.Get("error populate p2sh", err).Fatal()
		}
		jlog.Logf("Populated p2sh, blocks processed: %d\n", populateP2sh.BlocksProcessed)
	},
}
