package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LinkAccepted struct {
	RequestTxHash [32]byte
	TxHash        [32]byte
}

func (r *LinkAccepted) GetTopic() string {
	return db.TopicMemoLinkAccepted
}

func (r *LinkAccepted) GetShardSource() uint {
	return client.GenShardSource(r.RequestTxHash[:])
}

func (r *LinkAccepted) GetUid() []byte {
	return jutil.ByteReverse(r.RequestTxHash[:])
}

func (r *LinkAccepted) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		panic("invalid uid size for link accepted")
	}
	copy(r.RequestTxHash[:], jutil.ByteReverse(uid))
}

func (r *LinkAccepted) Serialize() []byte {
	return jutil.ByteReverse(r.TxHash[:])
}

func (r *LinkAccepted) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		panic("invalid data size for link accepted")
	}
	copy(r.TxHash[:], jutil.ByteReverse(data))
}
