package suite

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/queue"
	"github.com/memocash/server/test/run"
)

type Suite struct {
	Queue0 *run.Queue
	Queue1 *run.Queue
}

func (s *Suite) Start() error {
	s.Queue0 = run.NewQueue(queue.DefaultShard0Port)
	if err := s.Queue0.Start(); err != nil {
		return jerr.Get("error starting queue 0 server", err)
	}
	s.Queue1 = run.NewQueue(queue.DefaultShard1Port)
	if err := s.Queue0.Start(); err != nil {
		return jerr.Get("error starting queue 1 server", err)
	}
	return nil
}

func (s *Suite) Restart() error {
	err := s.End()
	if err != nil {
		return jerr.Get("error ending suite", err)
	}
	err = s.Start()
	if err != nil {
		return jerr.Get("error starting suite", err)
	}
	return nil
}

func (s *Suite) EndPrint() {
	err := s.End()
	if err != nil {
		jerr.Get("error ending suite", err).Print()
	}
}

func (s *Suite) End() error {
	s.Queue0.End()
	if s.Queue0.Error != nil {
		return jerr.Get("queue error", s.Queue0.Error)
	}
	return nil
}

func GetNewSuite() *Suite {
	return &Suite{}
}
