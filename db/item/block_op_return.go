package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
)

type BlockOpReturn struct {
	BlockHash []byte
	TxHash    []byte
	Index     uint32
	OpReturn  []byte
}

func (r BlockOpReturn) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(r.BlockHash),
		jutil.ByteReverse(r.TxHash),
		jutil.GetUint32Data(r.Index),
	)
}

func (r BlockOpReturn) GetShard() uint {
	return client.GetByteShard(r.TxHash)
}

func (r BlockOpReturn) GetTopic() string {
	return TopicBlockOpReturn
}

func (r BlockOpReturn) Serialize() []byte {
	return r.OpReturn
}

func (r *BlockOpReturn) SetUid(uid []byte) {
	if len(uid) != 68 {
		return
	}
	r.BlockHash = jutil.ByteReverse(uid[:32])
	r.TxHash = jutil.ByteReverse(uid[32:64])
	r.Index = jutil.GetUint32(uid[64:])
}

func (r *BlockOpReturn) Deserialize(data []byte) {
	r.OpReturn = data
}
