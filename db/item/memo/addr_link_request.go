package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

type AddrLinkRequest struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
}

func (r *AddrLinkRequest) GetTopic() string {
	return db.TopicMemoAddrLinkRequest
}

func (r *AddrLinkRequest) GetShardSource() uint {
	return client.GenShardSource(r.Addr[:])
}

func (r *AddrLinkRequest) GetUid() []byte {
	return jutil.CombineBytes(
		r.Addr[:],
		jutil.GetTimeByteNanoBig(r.Seen),
		jutil.ByteReverse(r.TxHash[:]),
	)
}

func (r *AddrLinkRequest) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(r.Addr[:], uid[:25])
	r.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(r.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (r *AddrLinkRequest) Serialize() []byte {
	return nil
}

func (r *AddrLinkRequest) Deserialize([]byte) {}
