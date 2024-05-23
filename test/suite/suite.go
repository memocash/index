package suite

import (
	"fmt"
	"github.com/memocash/index/db/store"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/test/run"
	"log"
	"os"
)

type Suite struct {
	Queue0 *run.Queue
	Queue1 *run.Queue
}

func (s *Suite) ClearData() error {
	if err := os.RemoveAll(store.GetDataDir()); err != nil {
		return fmt.Errorf("error removing store data directory; %w", err)
	}
	return nil
}

func (s *Suite) Start() error {
	if err := s.ClearData(); err != nil {
		return fmt.Errorf("error clearing data when starting suite; %w", err)
	}
	shards := config.GetQueueShards()
	if len(shards) != 2 {
		return fmt.Errorf("expected 2 shards, got %d", len(shards))
	}
	s.Queue0 = run.NewQueue(shards[0].Port, 0)
	if err := s.Queue0.Start(); err != nil {
		return fmt.Errorf("error starting queue 0 server; %w", err)
	}
	s.Queue1 = run.NewQueue(shards[1].Port, 1)
	if err := s.Queue1.Start(); err != nil {
		return fmt.Errorf("error starting queue 1 server; %w", err)
	}
	return nil
}

func (s *Suite) Restart() error {
	err := s.End()
	if err != nil {
		return fmt.Errorf("error ending suite; %w", err)
	}
	err = s.Start()
	if err != nil {
		return fmt.Errorf("error starting suite; %w", err)
	}
	return nil
}

func (s *Suite) EndPrint() {
	err := s.End()
	if err != nil {
		log.Printf("error ending suite; %v", err)
	}
}

func (s *Suite) End() error {
	if s.Queue0 != nil {
		if s.Queue0.End(); s.Queue0.Error != nil {
			return fmt.Errorf("server test suite queue0 error; %w", s.Queue0.Error)
		}
	}
	if s.Queue1 != nil {
		if s.Queue1.End(); s.Queue1.Error != nil {
			return fmt.Errorf("server test suite queue1 error; %w", s.Queue1.Error)
		}
	}
	return nil
}

func GetNewSuite() *Suite {
	return &Suite{}
}
