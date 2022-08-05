package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type MemoLikeTip struct {
	PostTxHash []byte
	LikeTxHash []byte
	Tip        int64
}

func (n MemoLikeTip) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(n.PostTxHash),
		jutil.ByteReverse(n.LikeTxHash),
	)
}

func (n MemoLikeTip) GetShard() uint {
	return client.GetByteShard(n.PostTxHash)
}

func (n MemoLikeTip) GetTopic() string {
	return TopicMemoLikeTip
}

func (n MemoLikeTip) Serialize() []byte {
	return jutil.GetInt64Data(n.Tip)
}

func (n *MemoLikeTip) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength*2 {
		panic("invalid uid size for memo like tip")
	}
	n.PostTxHash = jutil.ByteReverse(uid[:32])
	n.LikeTxHash = jutil.ByteReverse(uid[32:64])
}

func (n *MemoLikeTip) Deserialize(data []byte) {
	if len(data) != memo.Int8Size {
		panic("invalid data size for memo like tip")
	}
	n.Tip = jutil.GetInt64(data)
}
