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

type LockHeightFollowed struct {
	FollowLockHash []byte
	Height         int64
	TxHash         []byte
	LockHash       []byte
	Unfollow       bool
}

func (f LockHeightFollowed) GetUid() []byte {
	return jutil.CombineBytes(
		f.FollowLockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(f.Height)),
		jutil.ByteReverse(f.TxHash),
	)
}

func (f LockHeightFollowed) GetShard() uint {
	return client.GetByteShard(f.FollowLockHash)
}

func (f LockHeightFollowed) GetTopic() string {
	return db.TopicLockMemoFollowed
}

func (f LockHeightFollowed) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		f.LockHash,
	)
}

func (f *LockHeightFollowed) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	f.FollowLockHash = uid[:32]
	f.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	f.TxHash = jutil.ByteReverse(uid[40:72])
}

func (f *LockHeightFollowed) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+1 {
		return
	}
	f.Unfollow = data[0] == 1
	f.LockHash = data[1 : memo.TxHashLength+1]
}

func GetLockHeightFolloweds(ctx context.Context, followLockHashes [][]byte) ([]*LockHeightFollowed, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, followLockHash := range followLockHashes {
		shard := client.GetByteShard32(followLockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], followLockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockFolloweds []*LockHeightFollowed
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
			var lockFollowed = new(LockHeightFollowed)
			db.Set(lockFollowed, msg)
			lockFolloweds = append(lockFolloweds, lockFollowed)
		}
	}
	return lockFolloweds, nil
}

func GetLockHeightFollowedsSingle(ctx context.Context, followLockHash []byte, start int64) ([]*LockHeightFollowed, error) {
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
	var lockFolloweds = make([]*LockHeightFollowed, len(dbClient.Messages))
	for i := range dbClient.Messages {
		lockFolloweds[i] = new(LockHeightFollowed)
		db.Set(lockFolloweds[i], dbClient.Messages[i])
	}
	return lockFolloweds, nil
}

func RemoveLockHeightFollowed(lockFollow *LockHeightFollowed) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockFollow.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicLockMemoFollowed, [][]byte{lockFollow.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo followed", err)
	}
	return nil
}

func ListenLockHeightFolloweds(ctx context.Context, followLockHashes [][]byte) (chan *LockHeightFollowed, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, followLockHash := range followLockHashes {
		shard := client.GetByteShard32(followLockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], followLockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockFollowedChan = make(chan *LockHeightFollowed)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(lockFollowedChan)
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
				var lockFollowed = new(LockHeightFollowed)
				db.Set(lockFollowed, *msg)
				lockFollowedChan <- lockFollowed
			}
			cancelCtx.Cancel()
		}()
	}
	return lockFollowedChan, nil
}
