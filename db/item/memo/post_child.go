package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type PostChild struct {
	PostTxHash  [32]byte
	ChildTxHash [32]byte
}

func (c *PostChild) GetTopic() string {
	return db.TopicMemoPostChild
}

func (c *PostChild) GetShardSource() uint {
	return client.GenShardSource(c.PostTxHash[:])
}

func (c *PostChild) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(c.PostTxHash[:]),
		jutil.ByteReverse(c.ChildTxHash[:]),
	)
}

func (c *PostChild) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength*2 {
		return
	}
	copy(c.PostTxHash[:], jutil.ByteReverse(uid[:32]))
	copy(c.ChildTxHash[:], jutil.ByteReverse(uid[32:]))
}

func (c *PostChild) Serialize() []byte {
	return nil
}

func (c *PostChild) Deserialize([]byte) {}

func GetPostChildren(ctx context.Context, postTxHash [32]byte) ([]*PostChild, error) {
	shardConfig := config.GetShardConfig(db.GetShardIdFromByte32(postTxHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Context:  ctx,
		Topic:    db.TopicMemoPostChild,
		Prefixes: [][]byte{jutil.ByteReverse(postTxHash[:])},
	}); err != nil {
		return nil, fmt.Errorf("error getting client message memo post children; %w", err)
	}
	var postChildren = make([]*PostChild, len(dbClient.Messages))
	for i := range dbClient.Messages {
		postChildren[i] = new(PostChild)
		db.Set(postChildren[i], dbClient.Messages[i])
	}
	return postChildren, nil
}

func ListenPostChildren(ctx context.Context, postTxHashes [][32]byte) (chan *PostChild, error) {
	if len(postTxHashes) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range postTxHashes {
		shard := client.GenShardSource32(postTxHashes[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHashes[i][:]))
	}
	shardConfigs := config.GetQueueShards()
	var postChildChan = make(chan *PostChild)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(postChildChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoPostChild, prefixes)
		if err != nil {
			return nil, fmt.Errorf("error listening to db memo post child by prefix; %w", err)
		}
		go func() {
			for msg := range chanMessage {
				var postChild = new(PostChild)
				db.Set(postChild, *msg)
				postChildChan <- postChild
			}
			cancelCtx.Cancel()
		}()
	}
	return postChildChan, nil
}
