package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type MemoPostChild struct {
	PostTxHash  []byte
	ChildTxHash []byte
}

func (p MemoPostChild) GetUid() []byte {
	return jutil.ByteReverse(p.PostTxHash)
}

func (p MemoPostChild) GetShard() uint {
	return client.GetByteShard(p.PostTxHash)
}

func (p MemoPostChild) GetTopic() string {
	return TopicMemoPostChild
}

func (p MemoPostChild) Serialize() []byte {
	return jutil.ByteReverse(p.ChildTxHash)
}

func (p *MemoPostChild) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	p.PostTxHash = jutil.ByteReverse(uid)
}

func (p *MemoPostChild) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		return
	}
	p.ChildTxHash = jutil.ByteReverse(data)
}
