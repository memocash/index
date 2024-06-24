package debug

import (
	"github.com/memocash/index/test/suite"
	"github.com/memocash/index/test/tasks"
	"github.com/spf13/cobra"
	"log"
)

func getTestCommand() *cobra.Command {
	var testCmd = &cobra.Command{
		Use:   "test",
		Short: "Run tests",
	}
	for _, tst := range tasks.GetTests() {
		t := tst
		var cmd = &cobra.Command{
			Use: t.Name,
			RunE: func(c *cobra.Command, args []string) error {
				if err := suite.Run(&t, args); err != nil {
					log.Fatalf("fatal error running test; %v", err)
				}
				log.Printf("Suite (single) %s success!\n", t.Name)
				return nil
			},
		}
		testCmd.AddCommand(cmd)
	}
	return testCmd
}
