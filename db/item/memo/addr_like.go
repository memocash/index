package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

type AddrLike struct {
	Addr       [25]byte
	Seen       time.Time
	LikeTxHash [32]byte
	PostTxHash [32]byte
}

func (l *AddrLike) GetTopic() string {
	return db.TopicMemoAddrLike
}

func (l *AddrLike) GetShard() uint {
	return client.GetByteShard(l.Addr[:])
}

func (l *AddrLike) GetUid() []byte {
	return jutil.CombineBytes(
		l.Addr[:],
		jutil.GetTimeByteBig(l.Seen),
		jutil.ByteReverse(l.LikeTxHash[:]),
	)
}

func (l *AddrLike) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(l.Addr[:], uid[:25])
	l.Seen = jutil.GetByteTimeBig(uid[25:33])
	copy(l.LikeTxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (l *AddrLike) Serialize() []byte {
	return l.PostTxHash[:]
}

func (l *AddrLike) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		return
	}
	copy(l.PostTxHash[:], data)
}
