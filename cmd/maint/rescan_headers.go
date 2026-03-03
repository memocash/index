package maint

import (
	"log"

	"github.com/memocash/index/ref/cluster/lead"
	"github.com/spf13/cobra"
)

var rescanHeadersCmd = &cobra.Command{
	Use:   "rescan-headers",
	Short: "Rescan block headers from genesis, overwriting existing height/hash mappings",
	Run: func(c *cobra.Command, args []string) {
		scanHeaders := lead.NewScanHeaders()
		scanHeaders.Rescan = true
		log.Println("Starting full header rescan from genesis...")
		if err := scanHeaders.Run(); err != nil {
			log.Fatalf("error rescanning headers; %v", err)
		}
	},
}
