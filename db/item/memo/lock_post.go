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

type LockPost struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
}

func (p LockPost) GetUid() []byte {
	return jutil.CombineBytes(
		p.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash),
	)
}

func (p LockPost) GetShard() uint {
	return client.GetByteShard(p.LockHash)
}

func (p LockPost) GetTopic() string {
	return db.TopicLockMemoPost
}

func (p LockPost) Serialize() []byte {
	return nil
}

func (p *LockPost) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	p.LockHash = uid[:32]
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	p.TxHash = jutil.ByteReverse(uid[40:72])
}

func (p *LockPost) Deserialize([]byte) {}

func GetLockPosts(ctx context.Context, lockHashes [][]byte) ([]*LockPost, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockPosts []*LockPost
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicLockMemoPost,
			Prefixes: lockHashPrefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db lock memo post by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var lockPost = new(LockPost)
			db.Set(lockPost, msg)
			lockPosts = append(lockPosts, lockPost)
		}
	}
	return lockPosts, nil
}

func RemoveLockPost(lockPost *LockPost) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockPost.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicLockMemoPost, [][]byte{lockPost.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo post", err)
	}
	return nil
}
