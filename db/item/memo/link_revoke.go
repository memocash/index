package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LinkRevoke struct {
	TxHash       [32]byte
	Addr         [25]byte
	AcceptTxHash [32]byte
	Message      string
}

func (r *LinkRevoke) GetTopic() string {
	return db.TopicMemoLinkRevoke
}

func (r *LinkRevoke) GetShardSource() uint {
	return client.GenShardSource(r.TxHash[:])
}

func (r *LinkRevoke) GetUid() []byte {
	return jutil.ByteReverse(r.TxHash[:])
}

func (r *LinkRevoke) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		panic("invalid uid size for link revoke")
	}
	copy(r.TxHash[:], jutil.ByteReverse(uid))
}

func (r *LinkRevoke) Serialize() []byte {
	return jutil.CombineBytes(
		r.Addr[:],
		jutil.ByteReverse(r.AcceptTxHash[:]),
		[]byte(r.Message),
	)
}

func (r *LinkRevoke) Deserialize(data []byte) {
	if len(data) < memo.AddressLength+memo.TxHashLength {
		panic("invalid data size for link revoke")
	}
	copy(r.Addr[:], data[:25])
	copy(r.AcceptTxHash[:], jutil.ByteReverse(data[25:57]))
	r.Message = string(data[57:])
}
