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

type LockMemoFollowed struct {
	FollowLockHash []byte
	Height         int64
	TxHash         []byte
	LockHash       []byte
	Unfollow       bool
}

func (n LockMemoFollowed) GetUid() []byte {
	return jutil.CombineBytes(
		n.FollowLockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n LockMemoFollowed) GetShard() uint {
	return client.GetByteShard(n.FollowLockHash)
}

func (n LockMemoFollowed) GetTopic() string {
	return db.TopicLockMemoFollowed
}

func (n LockMemoFollowed) Serialize() []byte {
	var unfollow byte
	if n.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		n.LockHash,
	)
}

func (n *LockMemoFollowed) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.FollowLockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *LockMemoFollowed) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+1 {
		return
	}
	n.Unfollow = data[0] == 1
	n.LockHash = data[1 : memo.TxHashLength+1]
}

func GetLockMemoFollowed(ctx context.Context, followLockHashes [][]byte) ([]*LockMemoFollowed, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, followLockHash := range followLockHashes {
		shard := client.GetByteShard32(followLockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], followLockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockMemoFolloweds []*LockMemoFollowed
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicLockMemoFollowed,
			Prefixes: lockHashPrefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db lock memo followed by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var lockMemoFollowed = new(LockMemoFollowed)
			db.Set(lockMemoFollowed, msg)
			lockMemoFolloweds = append(lockMemoFolloweds, lockMemoFollowed)
		}
	}
	return lockMemoFolloweds, nil
}

func GetLockMemoFolloweds(ctx context.Context, followLockHash []byte, start int64) ([]*LockMemoFollowed, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(followLockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startByte []byte
	if start != 0 {
		startByte = jutil.CombineBytes(followLockHash, jutil.ByteFlip(jutil.GetInt64DataBig(start)))
	} else {
		startByte = followLockHash
	}
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicLockMemoFollowed,
		Prefixes: [][]byte{followLockHash},
		Start:    startByte,
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo follow by prefix", err)
	}
	var memoFolloweds []*LockMemoFollowed
	for _, msg := range dbClient.Messages {
		var lockMemoFollow = new(LockMemoFollowed)
		db.Set(lockMemoFollow, msg)
		memoFolloweds = append(memoFolloweds, lockMemoFollow)
	}
	return memoFolloweds, nil
}

func RemoveLockMemoFollowed(lockMemoFollow *LockMemoFollowed) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockMemoFollow.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicLockMemoFollowed, [][]byte{lockMemoFollow.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo followed", err)
	}
	return nil
}

func ListenLockMemoFolloweds(ctx context.Context, followLockHashes [][]byte) (chan *LockMemoFollowed, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, followLockHash := range followLockHashes {
		shard := client.GetByteShard32(followLockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], followLockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockMemoFollowedChan = make(chan *LockMemoFollowed)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(lockMemoFollowedChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		chanMessage, err := client.NewClient(shardConfig.GetHost()).
			Listen(cancelCtx.Context, db.TopicLockMemoFollowed, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo followeds by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockMemoFollowed = new(LockMemoFollowed)
				db.Set(lockMemoFollowed, *msg)
				lockMemoFollowedChan <- lockMemoFollowed
			}
			cancelCtx.Cancel()
		}()
	}
	return lockMemoFollowedChan, nil
}
