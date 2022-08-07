package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type MemoPostParent struct {
	PostTxHash   []byte
	ParentTxHash []byte
}

func (p MemoPostParent) GetUid() []byte {
	return jutil.ByteReverse(p.PostTxHash)
}

func (p MemoPostParent) GetShard() uint {
	return client.GetByteShard(p.PostTxHash)
}

func (p MemoPostParent) GetTopic() string {
	return TopicMemoPostParent
}

func (p MemoPostParent) Serialize() []byte {
	return jutil.ByteReverse(p.ParentTxHash)
}

func (p *MemoPostParent) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	p.PostTxHash = jutil.ByteReverse(uid)
}

func (p *MemoPostParent) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		return
	}
	p.ParentTxHash = jutil.ByteReverse(data)
}
