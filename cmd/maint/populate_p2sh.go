package maint

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var populateP2shCmd = &cobra.Command{
	Use:   "populate-p2sh",
	Short: "populate-p2sh",
	Run: func(c *cobra.Command, args []string) {
		populateP2sh := maint.NewPopulateP2sh()
		jlog.Logf("Starting populate p2sh...\n")
		if err := populateP2sh.Populate(); err != nil {
			jerr.Get("error populate p2sh", err).Fatal()
		}
		jlog.Logf("Populated p2sh, blocks processed: %d\n", populateP2sh.BlocksProcessed)
	},
}
