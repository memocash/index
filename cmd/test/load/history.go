package load

import (
	"github.com/spf13/cobra"
	"log"
)

var historyCmd = &cobra.Command{
	Use: "history",
	Run: func(c *cobra.Command, args []string) {
		log.Println("load history command")
	},
}
