package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type AddrHeightLike struct {
	Addr       [25]byte
	Height     int64
	LikeTxHash [32]byte
	PostTxHash [32]byte
}

func (l *AddrHeightLike) GetTopic() string {
	return db.TopicMemoAddrHeightLike
}

func (l *AddrHeightLike) GetShard() uint {
	return client.GetByteShard(l.Addr[:])
}

func (l *AddrHeightLike) GetUid() []byte {
	return jutil.CombineBytes(
		l.Addr[:],
		jutil.ByteFlip(jutil.GetInt64DataBig(l.Height)),
		jutil.ByteReverse(l.LikeTxHash[:]),
	)
}

func (l *AddrHeightLike) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(l.Addr[:], uid[:25])
	l.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[25:33]))
	copy(l.LikeTxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (l *AddrHeightLike) Serialize() []byte {
	return l.PostTxHash[:]
}

func (l *AddrHeightLike) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		return
	}
	copy(l.PostTxHash[:], data)
}
