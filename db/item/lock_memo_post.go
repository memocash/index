package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockMemoPost struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
}

func (p LockMemoPost) GetUid() []byte {
	return jutil.CombineBytes(
		p.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash),
	)
}

func (p LockMemoPost) GetShard() uint {
	return client.GetByteShard(p.LockHash)
}

func (p LockMemoPost) GetTopic() string {
	return TopicLockMemoPost
}

func (p LockMemoPost) Serialize() []byte {
	return nil
}

func (p *LockMemoPost) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	p.LockHash = uid[:32]
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	p.TxHash = jutil.ByteReverse(uid[40:72])
}

func (p *LockMemoPost) Deserialize([]byte) {}

func GetLockMemoPosts(ctx context.Context, lockHashes [][]byte) ([]*LockMemoPost, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockMemoPosts []*LockMemoPost
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetWOpts(client.Opts{
			Topic:    TopicLockMemoPost,
			Prefixes: lockHashPrefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db lock memo post by prefix", err)
		}
		for _, msg := range db.Messages {
			var lockMemoPost = new(LockMemoPost)
			Set(lockMemoPost, msg)
			lockMemoPosts = append(lockMemoPosts, lockMemoPost)
		}
	}
	return lockMemoPosts, nil
}

func RemoveLockMemoPost(lockMemoPost *LockMemoPost) error {
	shardConfig := config.GetShardConfig(GetShard32(lockMemoPost.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicLockMemoPost, [][]byte{lockMemoPost.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo post", err)
	}
	return nil
}
