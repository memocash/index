package memo

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type PostParent struct {
	PostTxHash   [32]byte
	ParentTxHash [32]byte
}

func (p *PostParent) GetTopic() string {
	return db.TopicMemoPostParent
}

func (p *PostParent) GetShard() uint {
	return client.GetByteShard(p.PostTxHash[:])
}

func (p *PostParent) GetUid() []byte {
	return jutil.ByteReverse(p.PostTxHash[:])
}

func (p *PostParent) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(p.PostTxHash[:], jutil.ByteReverse(uid))
}

func (p *PostParent) Serialize() []byte {
	return jutil.ByteReverse(p.ParentTxHash[:])
}

func (p *PostParent) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		return
	}
	copy(p.ParentTxHash[:], jutil.ByteReverse(data))
}

func GetPostParent(ctx context.Context, postTxHash [32]byte) (*PostParent, error) {
	shardConfig := config.GetShardConfig(db.GetShardByte32(postTxHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Context: ctx,
		Topic:   db.TopicMemoPostParent,
		Uids:    [][]byte{jutil.ByteReverse(postTxHash[:])},
	}); err != nil {
		return nil, jerr.Get("error getting client message memo post parents", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, nil
	}
	var postParent = new(PostParent)
	db.Set(postParent, dbClient.Messages[0])
	return postParent, nil
}
