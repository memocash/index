package chain

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type TxOutput struct {
	TxHash     [32]byte
	Index      uint32
	Value      int64
	LockScript []byte
}

func (t *TxOutput) GetUid() []byte {
	return GetTxOutputUid(t.TxHash, t.Index)
}

func (t *TxOutput) GetShard() uint {
	return client.GetByteShard(t.TxHash[:])
}

func (t *TxOutput) GetTopic() string {
	return db.TopicTxOutput
}

func (t *TxOutput) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.GetInt64Data(t.Value),
		t.LockScript,
	)
}

func (t *TxOutput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	copy(t.TxHash[:], jutil.ByteReverse(uid[:32]))
	t.Index = jutil.GetUint32(uid[32:36])
}

func (t *TxOutput) Deserialize(data []byte) {
	if len(data) < 8 {
		return
	}
	t.Value = jutil.GetInt64(data[:8])
	t.LockScript = data[8:]
}

func GetTxOutputUid(txHash [32]byte, index uint32) []byte {
	return db.GetTxHashIndexUid(txHash[:], index)
}
