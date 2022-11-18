package chain

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type Tx struct {
	TxHash   [32]byte
	Version  int32
	LockTime uint32
}

func (t *Tx) GetTopic() string {
	return db.TopicTx
}

func (t *Tx) GetShard() uint {
	return client.GetByteShard(t.TxHash[:])
}

func (t *Tx) GetUid() []byte {
	return t.TxHash[:]
}

func (t *Tx) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	copy(t.TxHash[:], uid)
}

func (t *Tx) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.GetInt32Data(t.Version),
		jutil.GetUint32Data(t.LockTime),
	)
}

func (t *Tx) Deserialize(data []byte) {
	if len(data) < 8 {
		return
	}
	t.Version = jutil.GetInt32(data[:4])
	t.LockTime = jutil.GetUint32(data[4:8])
}
