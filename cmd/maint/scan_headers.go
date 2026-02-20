package maint

import (
	"log"

	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var scanHeadersCmd = &cobra.Command{
	Use:   "scan-headers",
	Short: "Scan block headers from genesis and store height/hash mappings",
	Run: func(c *cobra.Command, args []string) {
		scanHeaders := maint.NewScanHeaders()
		log.Println("Starting header scan from genesis...")
		if err := scanHeaders.Run(); err != nil {
			log.Fatalf("error scanning headers; %v", err)
		}
	},
}
