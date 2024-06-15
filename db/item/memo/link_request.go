package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LinkRequest struct {
	TxHash     [32]byte
	ChildAddr  [25]byte
	ParentAddr [25]byte
	Message    string
}

func (r *LinkRequest) GetTopic() string {
	return db.TopicMemoLinkRequest
}

func (r *LinkRequest) GetShardSource() uint {
	return client.GenShardSource(r.TxHash[:])
}

func (r *LinkRequest) GetUid() []byte {
	return jutil.ByteReverse(r.TxHash[:])
}

func (r *LinkRequest) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		panic("invalid uid size for link request")
	}
	copy(r.TxHash[:], jutil.ByteReverse(uid))
}

func (r *LinkRequest) Serialize() []byte {
	return jutil.CombineBytes(
		r.ChildAddr[:],
		r.ParentAddr[:],
		[]byte(r.Message),
	)
}

func (r *LinkRequest) Deserialize(data []byte) {
	if len(data) < memo.AddressLength*2 {
		panic("invalid data size for link request")
	}
	copy(r.ChildAddr[:], data[:25])
	copy(r.ParentAddr[:], data[25:50])
	r.Message = string(data[50:])
}
