package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/admin/client/peer"
	"github.com/spf13/cobra"
)

var loopingEnableCmd = &cobra.Command{
	Use: "looping-enable",
	Run: func(cmd *cobra.Command, args []string) {
		loopingToggle := peer.NewLoopingToggle()
		if err := loopingToggle.Enable(); err != nil {
			jerr.Get("fatal error enabling looping", err).Fatal()
		}
		jlog.Logf("loopingToggle.Message: %s\n", loopingToggle.Message)
	},
}

var loopingDisableCmd = &cobra.Command{
	Use: "looping-disable",
	Run: func(cmd *cobra.Command, args []string) {
		loopingToggle := peer.NewLoopingToggle()
		if err := loopingToggle.Disable(); err != nil {
			jerr.Get("fatal error disabling looping", err).Fatal()
		}
		jlog.Logf("loopingToggle.Message: %s\n", loopingToggle.Message)
	},
}
