package peer

import (
	"github.com/memocash/index/admin/client/peer"
	"github.com/spf13/cobra"
	"log"
)

var loopingEnableCmd = &cobra.Command{
	Use: "looping-enable",
	Run: func(cmd *cobra.Command, args []string) {
		loopingToggle := peer.NewLoopingToggle()
		if err := loopingToggle.Enable(); err != nil {
			log.Fatalf("fatal error enabling looping; %v", err)
		}
		log.Printf("loopingToggle.Message: %s\n", loopingToggle.Message)
	},
}

var loopingDisableCmd = &cobra.Command{
	Use: "looping-disable",
	Run: func(cmd *cobra.Command, args []string) {
		loopingToggle := peer.NewLoopingToggle()
		if err := loopingToggle.Disable(); err != nil {
			log.Fatalf("fatal error disabling looping; %v", err)
		}
		log.Printf("loopingToggle.Message: %s\n", loopingToggle.Message)
	},
}
