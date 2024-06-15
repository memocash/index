package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LinkRevoked struct {
	AcceptTxHash [32]byte
	TxHash       [32]byte
}

func (r *LinkRevoked) GetTopic() string {
	return db.TopicMemoLinkRevoked
}

func (r *LinkRevoked) GetShardSource() uint {
	return client.GenShardSource(r.AcceptTxHash[:])
}

func (r *LinkRevoked) GetUid() []byte {
	return jutil.ByteReverse(r.AcceptTxHash[:])
}

func (r *LinkRevoked) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		panic("invalid uid size for link revoked")
	}
	copy(r.AcceptTxHash[:], jutil.ByteReverse(uid))
}

func (r *LinkRevoked) Serialize() []byte {
	return jutil.ByteReverse(r.TxHash[:])
}

func (r *LinkRevoked) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		panic("invalid data size for link revoked")
	}
	copy(r.TxHash[:], jutil.ByteReverse(data))
}
