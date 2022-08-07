package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
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

func GetMemoPostParent(ctx context.Context, postTxHash []byte) (*MemoPostParent, error) {
	shardConfig := config.GetShardConfig(GetShardByte32(postTxHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Context: ctx,
		Topic:   TopicMemoPostParent,
		Uids:    [][]byte{jutil.ByteReverse(postTxHash)},
	}); err != nil {
		return nil, jerr.Get("error getting client message memo post parents", err)
	}
	if len(db.Messages) == 0 {
		return nil, nil
	}
	var memoPostParent = new(MemoPostParent)
	Set(memoPostParent, db.Messages[0])
	return memoPostParent, nil
}
