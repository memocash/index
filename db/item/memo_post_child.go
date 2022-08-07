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
	return jutil.CombineBytes(
		jutil.ByteReverse(p.PostTxHash),
		jutil.ByteReverse(p.ChildTxHash),
	)
}

func (p MemoPostChild) GetShard() uint {
	return client.GetByteShard(p.PostTxHash)
}

func (p MemoPostChild) GetTopic() string {
	return TopicMemoPostChild
}

func (p MemoPostChild) Serialize() []byte {
	return nil
}

func (p *MemoPostChild) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength*2 {
		return
	}
	p.PostTxHash = jutil.ByteReverse(uid[:32])
	p.ChildTxHash = jutil.ByteReverse(uid[32:])
}

func (p *MemoPostChild) Deserialize([]byte) {}
