package chain

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type BlockInfo struct {
	BlockHash [32]byte
	Size      int64
	TxCount   int
}

func (b *BlockInfo) GetTopic() string {
	return db.TopicChainBlockInfo
}

func (b *BlockInfo) GetShard() uint {
	return client.GetByteShard(b.BlockHash[:])
}

func (b *BlockInfo) GetUid() []byte {
	return b.BlockHash[:]
}

func (b *BlockInfo) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	copy(b.BlockHash[:], jutil.ByteReverse(uid[:32]))
}

func (b *BlockInfo) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.GetInt64Data(b.Size),
		jutil.GetIntData(b.TxCount),
	)
}

func (b *BlockInfo) Deserialize(data []byte) {
	if len(data) != 12 {
		return
	}
	b.Size = jutil.GetInt64(data[:8])
	b.TxCount = jutil.GetInt(data[8:12])
}
