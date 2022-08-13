package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
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
	return db.TopicMessage
}

func (t Message) Serialize() []byte {
	return jutil.CombineBytes(jutil.GetTimeByteNano(t.Created), []byte(t.Message))
}

func (t *Message) SetUid(uid []byte) {
	t.Id = jutil.GetUint(uid)
}

func (t *Message) Deserialize(data []byte) {
	if len(data) < 8 {
		return
	}
	t.Created = jutil.GetByteTimeNano(data[:8])
	t.Message = string(data[8:])
}

func GetMessage(id uint) (*Message, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(jutil.GetUintData(id)), config.GetQueueShards())
	queueClient := client.NewClient(shardConfig.GetHost())
	if err := queueClient.GetSingle(db.TopicMessage, jutil.GetUintData(id)); err != nil {
		return nil, jerr.Get("error getting single client message", err)
	}
	if len(queueClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of messages: %d", len(queueClient.Messages))
	}
	var message = new(Message)
	db.Set(message, queueClient.Messages[0])
	return message, nil
}
