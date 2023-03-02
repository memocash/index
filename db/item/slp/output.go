package slp

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Output struct {
	TxHash    [32]byte
	Index     uint32
	TokenHash [32]byte
	Quantity  uint64
}

func (o *Output) GetTopic() string {
	return db.TopicSlpOutput
}

func (o *Output) GetShard() uint {
	return client.GetByteShard(o.TxHash[:])
}

func (o *Output) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(o.TxHash[:]),
		jutil.GetUint32Data(o.Index),
	)
}

func (o *Output) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+4 {
		return
	}
	copy(o.TxHash[:32], jutil.ByteReverse(uid))
	o.Index = jutil.GetUint32(uid[32:])
}

func (o *Output) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(o.TokenHash[:]),
		jutil.GetUint64Data(o.Quantity),
	)
}

func (o *Output) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+8 {
		return
	}
	copy(o.TokenHash[:], jutil.ByteReverse(data[:32]))
	o.Quantity = jutil.GetUint64(data[32:])
}
