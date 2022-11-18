package chain

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type BlockTx struct {
	BlockHash [32]byte
	Index     uint32
	TxHash    [32]byte
}

func (b *BlockTx) GetTopic() string {
	return db.TopicBlockTx
}

func (b *BlockTx) GetShard() uint {
	return client.GetByteShard(b.BlockHash[:])
}

func (b *BlockTx) GetUid() []byte {
	return GetBlockTxUid(b.BlockHash, b.Index)
}

func (b *BlockTx) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	copy(b.BlockHash[:], jutil.ByteReverse(uid[:32]))
	b.Index = jutil.GetUint32Big(uid[32:36])
}

func (b *BlockTx) Serialize() []byte {
	return b.TxHash[:]
}

func (b *BlockTx) Deserialize(data []byte) {
	if len(data) != 32 {
		return
	}
	copy(b.TxHash[:], jutil.ByteReverse(data[:32]))
}

func GetBlockTxUid(blockHash [32]byte, index uint32) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(blockHash[:]), jutil.GetUint32DataBig(index))
}
