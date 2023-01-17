package status

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
)

type Status struct {
	Topic  string
	Shard  uint
	Status *item.ProcessStatus
}

func (s *Status) error(err error) {
	jerr.Get("error saving tx queue", err).Print()
}

func (s *Status) SetStatus(status []byte) error {
	if err := s.setProcessStatus(); err != nil {
		return jerr.Get("error setting process status", err)
	}
	s.Status.Status = status
	if err := s.Status.Save(); err != nil {
		return jerr.Get("error saving status", err)
	}
	return nil
}

func (s *Status) GetStatus() ([]byte, error) {
	if err := s.setProcessStatus(); err != nil {
		return nil, jerr.Get("error setting process status", err)
	}
	return s.Status.Status, nil
}

func (s *Status) setProcessStatus() error {
	if s.Status != nil {
		return nil
	}
	processStatus, err := item.GetProcessStatus(s.Shard, s.Topic)
	if err != nil && !client.IsMessageNotSetError(err) {
		return jerr.Get("error getting process status from db", err)
	}
	if processStatus != nil {
		s.Status = processStatus
	} else {
		s.Status = item.NewProcessStatus(s.Shard, s.Topic)
	}
	return nil
}

func NewStatus(topic string, shard uint) *Status {
	return &Status{
		Topic: topic,
		Shard: shard,
	}
}
