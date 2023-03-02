package slp

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Mint struct {
	TxHash     [32]byte
	TokenHash  [32]byte
	BatonIndex uint32
	Quantity   uint64
}

func (m *Mint) GetTopic() string {
	return db.TopicSlpMint
}

func (m *Mint) GetShard() uint {
	return client.GetByteShard(m.TxHash[:])
}

func (m *Mint) GetUid() []byte {
	return jutil.ByteReverse(m.TxHash[:])
}

func (m *Mint) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(m.TxHash[:], jutil.ByteReverse(uid))
}

func (m *Mint) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(m.TokenHash[:]),
		jutil.GetUint32Data(m.BatonIndex),
		jutil.GetUint64Data(m.Quantity),
	)
}

func (m *Mint) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+4+8 {
		return
	}
	copy(m.TokenHash[:], jutil.ByteReverse(data[:memo.TxHashLength]))
	m.BatonIndex = jutil.GetUint32(data[memo.TxHashLength : memo.TxHashLength+4])
	m.Quantity = jutil.GetUint64(data[memo.TxHashLength+4:])
}
