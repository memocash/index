package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

type AddrLinkRequested struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
}

func (r *AddrLinkRequested) GetTopic() string {
	return db.TopicMemoAddrLinkRequested
}

func (r *AddrLinkRequested) GetShardSource() uint {
	return client.GenShardSource(r.Addr[:])
}

func (r *AddrLinkRequested) GetUid() []byte {
	return jutil.CombineBytes(
		r.Addr[:],
		jutil.GetTimeByteNanoBig(r.Seen),
		jutil.ByteReverse(r.TxHash[:]),
	)
}

func (r *AddrLinkRequested) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(r.Addr[:], uid[:25])
	r.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(r.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (r *AddrLinkRequested) Serialize() []byte {
	return nil
}

func (r *AddrLinkRequested) Deserialize([]byte) {}
