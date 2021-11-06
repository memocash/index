package suite

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/store"
	"github.com/memocash/server/ref/config"
	"github.com/memocash/server/test/run"
	"os"
)

type Suite struct {
	Queue0 *run.Queue
	Queue1 *run.Queue
}

func (s *Suite) ClearData() error {
	if err := os.RemoveAll(store.GetDataDir()); err != nil {
		return jerr.Get("error removing store data directory", err)
	}
	return nil
}

func (s *Suite) Start() error {
	if err := s.ClearData(); err != nil {
		return jerr.Get("error clearing data when starting suite", err)
	}
	s.Queue0 = run.NewQueue(config.DefaultShard0Port, 0)
	if err := s.Queue0.Start(); err != nil {
		return jerr.Get("error starting queue 0 server", err)
	}
	s.Queue1 = run.NewQueue(config.DefaultShard1Port, 1)
	if err := s.Queue1.Start(); err != nil {
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
	if s.Queue0 != nil {
		if s.Queue0.End(); s.Queue0.Error != nil {
			return jerr.Get("server test suite queue0 error", s.Queue0.Error)
		}
	}
	if s.Queue1 != nil {
		if s.Queue1.End(); s.Queue1.Error != nil {
			return jerr.Get("server test suite queue1 error", s.Queue1.Error)
		}
	}
	return nil
}

func GetNewSuite() *Suite {
	return &Suite{}
}
