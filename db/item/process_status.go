package item

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

const (
	ProcessStatusPopulateP2sh       = "populate-p2sh"
	ProcessStatusPopulateAddr       = "populate-addr"
	ProcessStatusPopulateAddrInputs = "populate-addr-inputs"
)

type ProcessStatus struct {
	Name   string
	Shard  uint
	Status []byte
}

func (s *ProcessStatus) GetTopic() string {
	return db.TopicProcessStatus
}

func (s *ProcessStatus) GetShardSource() uint {
	return s.Shard
}

func (s *ProcessStatus) GetUid() []byte {
	return []byte(s.Name)
}

func (s *ProcessStatus) SetUid(uid []byte) {
	s.Name = string(uid)
}

func (s *ProcessStatus) Serialize() []byte {
	return s.Status
}

func (s *ProcessStatus) Deserialize(data []byte) {
	s.Status = data
}

func (s *ProcessStatus) Save() error {
	if err := db.Save([]db.Object{s}); err != nil {
		return fmt.Errorf("error saving process status; %w", err)
	}
	return nil
}

func NewProcessStatus(shard uint, name string) *ProcessStatus {
	return &ProcessStatus{
		Name:  name,
		Shard: shard,
	}
}

func GetProcessStatus(ctx context.Context, shard uint, name string) (*ProcessStatus, error) {
	shardConfig := config.GetShardConfig(uint32(shard), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(ctx, db.TopicProcessStatus, []byte(name)); err != nil {
		return nil, fmt.Errorf("error getting db message process status; %w", err)
	}
	if len(dbClient.Messages) == 0 || len(dbClient.Messages[0].Uid) == 0 {
		return nil, fmt.Errorf("error status not found; %w", client.MessageNotSetError)
	}
	var processStatus = new(ProcessStatus)
	db.Set(processStatus, dbClient.Messages[0])
	processStatus.Shard = shard
	return processStatus, nil
}
