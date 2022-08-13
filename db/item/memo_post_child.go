package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
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
	return db.TopicMemoPostChild
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
	shardConfig := config.GetShardConfig(db.GetShardByte32(postTxHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Context:  ctx,
		Topic:    db.TopicMemoPostChild,
		Prefixes: [][]byte{jutil.ByteReverse(postTxHash)},
	}); err != nil {
		return nil, jerr.Get("error getting client message memo post children", err)
	}
	var memoPostChildren = make([]*MemoPostChild, len(dbClient.Messages))
	for i := range dbClient.Messages {
		memoPostChildren[i] = new(MemoPostChild)
		db.Set(memoPostChildren[i], dbClient.Messages[i])
	}
	return memoPostChildren, nil
}
