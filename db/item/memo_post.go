package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoPost struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Post     string
}

func (n MemoPost) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n MemoPost) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n MemoPost) GetTopic() string {
	return TopicMemoPost
}

func (n MemoPost) Serialize() []byte {
	return []byte(n.Post)
}

func (n *MemoPost) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoPost) Deserialize(data []byte) {
	n.Post = string(data)
}

func GetMemoPost(ctx context.Context, lockHashes [][]byte) ([]*MemoPost, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var memoPosts []*MemoPost
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetWOpts(client.Opts{
			Topic:    TopicMemoPost,
			Prefixes: lockHashPrefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db memo post by prefix", err)
		}
		for _, msg := range db.Messages {
			var memoPost = new(MemoPost)
			Set(memoPost, msg)
			memoPosts = append(memoPosts, memoPost)
		}
	}
	return memoPosts, nil
}

func RemoveMemoPost(memoPost *MemoPost) error {
	shardConfig := config.GetShardConfig(GetShard32(memoPost.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicMemoPost, [][]byte{memoPost.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic memo post", err)
	}
	return nil
}

func ListenMemoPosts(ctx context.Context, lockHashes [][]byte) (chan *MemoPost, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var memoPostChan = make(chan *MemoPost)
	cancelCtx := NewCancelContext(ctx, func() {
		close(memoPostChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(cancelCtx.Context, TopicMemoPost, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo posts by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var memoPost = new(MemoPost)
				Set(memoPost, *msg)
				memoPostChan <- memoPost
			}
			cancelCtx.Cancel()
		}()
	}
	return memoPostChan, nil
}
