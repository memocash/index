package test

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/cmd/test/fund"
	"github.com/memocash/index/test/suite"
	"github.com/memocash/index/test/tasks"
	"github.com/spf13/cobra"
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
						jerr.Get("fatal error running test", err).Fatal()
					}
					jlog.Logf("Suite (single) %s success!\n", t.Name)
					return nil
				},
			}
			testCmd.AddCommand(cmd)
		}
	}
	testCmd.AddCommand(fund.GetCommand())
	return testCmd
}
