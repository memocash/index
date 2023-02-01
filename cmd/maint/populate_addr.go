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
		populateAddr := maint.NewPopulateAddr(false)
		jlog.Logf("Starting populate addr outputs...\n")
		if err := populateAddr.Populate(restart); err != nil {
			jerr.Get("error populate addr outputs", err).Fatal()
		}
		jlog.Logf("Populated addr outputs completed. Checked: %d, saved: %d.\n", populateAddr.Checked, populateAddr.Saved)
	},
}

var populateAddrInputsCmd = &cobra.Command{
	Use:   "populate-addr-inputs",
	Short: "populate-addr-inputs",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populateAddr := maint.NewPopulateAddr(true)
		jlog.Logf("Starting populate addr inputs...\n")
		if err := populateAddr.Populate(restart); err != nil {
			jerr.Get("error populate addr inputs", err).Fatal()
		}
		jlog.Logf("Populated addr inputs completed. Checked: %d, saved: %d.\n", populateAddr.Checked, populateAddr.Saved)
	},
}
