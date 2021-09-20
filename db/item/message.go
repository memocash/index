package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
	"time"
)

type Message struct {
	Id      uint
	Message string
	Created time.Time
}

func (t Message) GetUid() []byte {
	return jutil.GetUintData(t.Id)
}

func (t Message) GetShard() uint {
	return client.GetByteShard(t.GetUid())
}

func (t Message) GetTopic() string {
	return TopicMessage
}

func (t Message) Serialize() []byte {
	return jutil.CombineBytes(jutil.GetTimeByte(t.Created), []byte(t.Message))
}

func (t *Message) SetUid(uid []byte) {
	t.Id = jutil.GetUint(uid)
}

func (t *Message) Deserialize(data []byte) {
	if len(data) < 8 {
		return
	}
	t.Created = jutil.GetByteTime(data[:8])
	t.Message = string(data[8:])
}

func GetMessage(id uint) (*Message, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(jutil.GetUintData(id)), config.GetQueueShards())
	queueClient := client.NewClient(shardConfig.GetHost())
	if err := queueClient.GetSingle(TopicMessage, jutil.GetUintData(id)); err != nil {
		return nil, jerr.Get("error getting single client message", err)
	}
	if len(queueClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of messages: %d", len(queueClient.Messages))
	}
	var message = new(Message)
	message.SetUid(queueClient.Messages[0].Uid)
	message.Deserialize(queueClient.Messages[0].Message)
	return message, nil
}
