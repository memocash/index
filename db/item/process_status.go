package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
)

const (
	ProcessStatusTopicBlocks       = "blocks"
	ProcessStatusTopicBlockHeights = "block-heights"
	ProcessStatusTopicBlockTxes    = "block-txes"
)

type ProcessStatus struct {
	Name   string
	Height int64
	Shard  uint
}

func (s ProcessStatus) GetUid() []byte {
	return []byte(s.Name)
}

func (s ProcessStatus) GetShard() uint {
	return s.Shard
}

func (s ProcessStatus) GetTopic() string {
	return TopicProcessStatus
}

func (s ProcessStatus) Serialize() []byte {
	return jutil.GetInt64DataBig(s.Height)
}

func (s *ProcessStatus) SetUid(uid []byte) {
	s.Name = string(uid)
}

func (s *ProcessStatus) Deserialize(data []byte) {
	s.Height = jutil.GetInt64Big(data)
}

func (s *ProcessStatus) Save() error {
	err := Save([]Object{s})
	if err != nil {
		return jerr.Get("error saving process status", err)
	}
	return nil
}

func NewProcessStatus(shard uint, name string) *ProcessStatus {
	return &ProcessStatus{
		Name:  name,
		Shard: shard,
	}
}

func GetProcessStatus(shard uint, name string) (*ProcessStatus, error) {
	shardConfig := config.GetShardConfig(uint32(shard), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	err := db.GetSingle(TopicProcessStatus, []byte(name))
	if err != nil {
		return nil, jerr.Get("error getting db message process status", err)
	}
	if len(db.Messages) == 0 || len(db.Messages[0].Uid) == 0 {
		return nil, jerr.Get("error status not found", client.MessageNotSetError)
	}
	var processStatus = new(ProcessStatus)
	processStatus.SetUid(db.Messages[0].Uid)
	processStatus.Deserialize(db.Messages[0].Message)
	processStatus.Shard = shard
	return processStatus, nil
}
