package test

import (
	"github.com/memocash/index/cmd/test/fund"
	"github.com/memocash/index/cmd/test/graph"
	"github.com/memocash/index/test/suite"
	"github.com/memocash/index/test/tasks"
	"github.com/spf13/cobra"
	"log"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests",
}

var initCmd bool

func GetCommand() *cobra.Command {
	if !initCmd {
		initCmd = true
		for _, tst := range tasks.GetTests() {
			t := tst
			var cmd = &cobra.Command{
				Use: t.Name,
				RunE: func(c *cobra.Command, args []string) error {
					err := suite.Run(&t, args)
					if err != nil {
						log.Fatalf("fatal error running test; %v", err)
					}
					log.Printf("Suite (single) %s success!\n", t.Name)
					return nil
				},
			}
			testCmd.AddCommand(cmd)
		}
	}
	testCmd.AddCommand(
		saveTxCmd,
		itemDeleteCmd,
		fund.GetCommand(),
		graph.GetCommand(),
	)
	return testCmd
}
