package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LinkAccept struct {
	TxHash        [32]byte
	Addr          [25]byte
	RequestTxHash [32]byte
	Message       string
}

func (r *LinkAccept) GetTopic() string {
	return db.TopicMemoLinkAccept
}

func (r *LinkAccept) GetShardSource() uint {
	return client.GenShardSource(r.TxHash[:])
}

func (r *LinkAccept) GetUid() []byte {
	return jutil.ByteReverse(r.TxHash[:])
}

func (r *LinkAccept) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		panic("invalid uid size for link accept")
	}
	copy(r.TxHash[:], jutil.ByteReverse(uid))
}

func (r *LinkAccept) Serialize() []byte {
	return jutil.CombineBytes(
		r.Addr[:],
		jutil.ByteReverse(r.RequestTxHash[:]),
		[]byte(r.Message),
	)
}

func (r *LinkAccept) Deserialize(data []byte) {
	if len(data) < memo.AddressLength+memo.TxHashLength {
		panic("invalid data size for link accept")
	}
	copy(r.Addr[:], data[:25])
	copy(r.RequestTxHash[:], jutil.ByteReverse(data[25:57]))
	r.Message = string(data[57:])
}
