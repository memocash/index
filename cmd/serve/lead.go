package serve

import (
	"github.com/memocash/index/ref/cluster/lead"
	"github.com/spf13/cobra"
	"log"
)

var leadCmd = &cobra.Command{
	Use:   "lead",
	Short: "lead",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		p := lead.NewProcessor(verbose)
		log.Fatalf("fatal error running lead processor; %v", p.Run())
	},
}
