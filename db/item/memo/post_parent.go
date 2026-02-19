package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type PostParent struct {
	PostTxHash   [32]byte
	ParentTxHash [32]byte
}

func (p *PostParent) GetTopic() string {
	return db.TopicMemoPostParent
}

func (p *PostParent) GetShardSource() uint {
	return client.GenShardSource(p.PostTxHash[:])
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
	var postParent = &PostParent{PostTxHash: postTxHash}
	if err := db.GetItem(ctx, postParent); err != nil {
		return nil, fmt.Errorf("error getting client message memo post parent; %w", err)
	}
	return postParent, nil
}

func GetPostParents(ctx context.Context, postTxHashes [][32]byte) ([]*PostParent, error) {
	messages, err := db.GetSpecific(ctx, db.TopicMemoPostParent, db.ShardUidsTxHashes(postTxHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client message memo post parents; %w", err)
	}
	var postParents = make([]*PostParent, len(messages))
	for i := range messages {
		postParents[i] = new(PostParent)
		db.Set(postParents[i], messages[i])
	}
	return postParents, nil
}
