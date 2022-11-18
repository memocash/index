package chain

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type TxBlock struct {
	TxHash    [32]byte
	BlockHash [32]byte
	Index     uint32
}

func (b *TxBlock) GetTopic() string {
	return db.TopicTxBlock
}

func (b *TxBlock) GetShard() uint {
	return client.GetByteShard(b.TxHash[:])
}

func (b *TxBlock) GetUid() []byte {
	return GetTxBlockUid(b.TxHash, b.BlockHash)
}

func (b *TxBlock) SetUid(uid []byte) {
	if len(uid) != 64 {
		return
	}
	copy(b.TxHash[:], jutil.ByteReverse(uid[:32]))
	copy(b.BlockHash[:], jutil.ByteReverse(uid[32:64]))
}

func (b *TxBlock) Serialize() []byte {
	return nil
}

func (b *TxBlock) Deserialize([]byte) {}

func GetTxBlockUid(txHash [32]byte, blockHash [32]byte) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash[:]), jutil.ByteReverse(blockHash[:]))
}
