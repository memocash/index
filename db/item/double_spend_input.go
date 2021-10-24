package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
)

type DoubleSpendInput struct {
	TxHash []byte
	Index  uint32
}

func (i DoubleSpendInput) GetUid() []byte {
	return GetTxHashIndexUid(i.TxHash, i.Index)
}

func (i DoubleSpendInput) GetShard() uint {
	return client.GetByteShard(i.TxHash)
}

func (i DoubleSpendInput) GetTopic() string {
	return TopicDoubleSpendInput
}

func (i DoubleSpendInput) Serialize() []byte {
	return nil
}

func (i *DoubleSpendInput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	i.TxHash = jutil.ByteReverse(uid[:32])
	i.Index = jutil.GetUint32(uid[32:36])
}

func (i *DoubleSpendInput) Deserialize([]byte) {
	return
}
