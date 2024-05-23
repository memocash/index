package suite

import (
	"fmt"
	"log"
)

func Run(test *Test, args []string) error {
	s := GetNewSuite()
	err := s.Start()
	defer s.EndPrint()
	if err != nil {
		return fmt.Errorf("error starting new suite; %w", err)
	}
	log.Printf("Starting suite: %s...\n", test.Name)
	err = test.Test(&TestRequest{
		Suite: s,
		Args:  args,
	})
	if err != nil {
		return fmt.Errorf("error running suite %s; %w", test.Name, err)
	}
	return nil
}
