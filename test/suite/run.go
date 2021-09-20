package suite

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
)

func Run(test *Test, args []string) error {
	s := GetNewSuite()
	err := s.Start()
	defer s.EndPrint()
	if err != nil {
		return jerr.Get("error starting new suite", err)
	}
	jlog.Logf("Starting suite: %s...\n", test.Name)
	err = test.Test(&TestRequest{
		Suite: s,
		Args:  args,
	})
	if err != nil {
		return jerr.Getf(err, "error running suite %s", test.Name)
	}
	return nil
}
