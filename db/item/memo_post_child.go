package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
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

func GetMemoPostChildren(ctx context.Context, postTxHash []byte) ([]*MemoPostChild, error) {
	shardConfig := config.GetShardConfig(GetShardByte32(postTxHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Context:  ctx,
		Topic:    TopicMemoPostChild,
		Prefixes: [][]byte{jutil.ByteReverse(postTxHash)},
	}); err != nil {
		return nil, jerr.Get("error getting client message memo post children", err)
	}
	var memoPostChildren = make([]*MemoPostChild, len(db.Messages))
	for i := range db.Messages {
		memoPostChildren[i] = new(MemoPostChild)
		Set(memoPostChildren[i], db.Messages[i])
	}
	return memoPostChildren, nil
}
