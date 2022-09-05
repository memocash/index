package cluster

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/cluster/lead"
	"github.com/spf13/cobra"
)

var leadCmd = &cobra.Command{
	Use:   "lead",
	Short: "lead",
	Run: func(c *cobra.Command, args []string) {
		l := lead.NewLead()
		jerr.Get("fatal error running leader", l.Run()).Fatal()
	},
}
