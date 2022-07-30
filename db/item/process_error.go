package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
)

type ProcessError struct {
	TxHash []byte
	Error  string
}

func (e ProcessError) GetUid() []byte {
	return jutil.ByteReverse(e.TxHash)
}

func (e ProcessError) GetShard() uint {
	return client.GetByteShard(e.TxHash)
}

func (e ProcessError) GetTopic() string {
	return TopicProcessError
}

func (e ProcessError) Serialize() []byte {
	return []byte(e.Error)
}

func (e *ProcessError) SetUid(uid []byte) {
	e.TxHash = jutil.ByteReverse(uid)
}

func (e *ProcessError) Deserialize(data []byte) {
	e.Error = string(data)
}
