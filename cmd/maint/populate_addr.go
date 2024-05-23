package maint

import (
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
	"log"
)

var populateAddrOutputsCmd = &cobra.Command{
	Use:   "populate-addr-outputs",
	Short: "populate-addr-outputs",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populateAddr := maint.NewPopulateAddr(false)
		log.Printf("Starting populate addr outputs...\n")
		if err := populateAddr.Populate(restart); err != nil {
			log.Fatalf("error populate addr outputs; %v", err)
		}
		log.Printf("Populated addr outputs completed. Checked: %d, saved: %d.\n", populateAddr.Checked, populateAddr.Saved)
	},
}

var populateAddrInputsCmd = &cobra.Command{
	Use:   "populate-addr-inputs",
	Short: "populate-addr-inputs",
	Run: func(c *cobra.Command, args []string) {
		restart, _ := c.Flags().GetBool(FlagRestart)
		populateAddr := maint.NewPopulateAddr(true)
		log.Printf("Starting populate addr inputs...\n")
		if err := populateAddr.Populate(restart); err != nil {
			log.Fatalf("error populate addr inputs; %v", err)
		}
		log.Printf("Populated addr inputs completed. Checked: %d, saved: %d.\n", populateAddr.Checked, populateAddr.Saved)
	},
}
