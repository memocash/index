package status

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
)

type Status struct {
	Topic  string
	Shard  uint
	Status *item.ProcessStatus
}

func (s *Status) error(err error) {
	jerr.Get("error saving tx queue", err).Print()
}

func (s *Status) SetHeight(height int64) error {
	err := s.setProcessStatus()
	if err != nil {
		return jerr.Get("error setting process status", err)
	}
	s.Status.Height = height
	err = s.Status.Save()
	if err != nil {
		return jerr.Get("error saving status", err)
	}
	return nil
}

func (s *Status) GetHeight() (int64, error) {
	err := s.setProcessStatus()
	if err != nil {
		return 0, jerr.Get("error setting process status", err)
	}
	return s.Status.Height, nil
}

func (s *Status) setProcessStatus() error {
	if s.Status != nil {
		return nil
	}
	processStatus, err := item.GetProcessStatus(s.Shard, s.Topic)
	if err != nil && ! client.IsMessageNotSetError(err) {
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

