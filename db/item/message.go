package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
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
