package maint

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var populateAddrOutputsCmd = &cobra.Command{
	Use:   "populate-addr-outputs",
	Short: "populate-addr-outputs",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populateAddr := maint.NewPopulateAddr()
		jlog.Logf("Starting populate addr outputs...\n")
		if err := populateAddr.Populate(restart); err != nil {
			jerr.Get("error populate addr outputs", err).Fatal()
		}
		jlog.Logf("Populated addr outputs completed. Checked: %d, saved: %d.\n", populateAddr.Checked, populateAddr.Saved)
	},
}
