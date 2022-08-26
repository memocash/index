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

type LockHeightFollow struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Unfollow bool
	Follow   []byte
}

func (f LockHeightFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(f.Height)),
		jutil.ByteReverse(f.TxHash),
	)
}

func (f LockHeightFollow) GetShard() uint {
	return client.GetByteShard(f.LockHash)
}

func (f LockHeightFollow) GetTopic() string {
	return db.TopicMemoLockHeightFollow
}

func (f LockHeightFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		f.Follow,
	)
}

func (f *LockHeightFollow) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	f.LockHash = uid[:32]
	f.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	f.TxHash = jutil.ByteReverse(uid[40:72])
}

func (f *LockHeightFollow) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+1 {
		return
	}
	f.Unfollow = data[0] == 1
	f.Follow = data[1 : memo.TxHashLength+1]
}

func GetLockHeightFollows(ctx context.Context, lockHashes [][]byte) ([]*LockHeightFollow, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockFollows []*LockHeightFollow
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoLockHeightFollow,
			Prefixes: lockHashPrefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db lock memo follow by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var lockFollow = new(LockHeightFollow)
			db.Set(lockFollow, msg)
			lockFollows = append(lockFollows, lockFollow)
		}
	}
	return lockFollows, nil
}

func GetLockHeightFollowsSingle(ctx context.Context, lockHash []byte, start int64) ([]*LockHeightFollow, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startByte []byte
	if start != 0 {
		startByte = jutil.CombineBytes(lockHash, jutil.ByteFlip(jutil.GetInt64DataBig(start)))
	} else {
		startByte = lockHash
	}
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoLockHeightFollow,
		Prefixes: [][]byte{lockHash},
		Start:    startByte,
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo follow by prefix", err)
	}
	var lockFollows = make([]*LockHeightFollow, len(dbClient.Messages))
	for i := range dbClient.Messages {
		lockFollows[i] = new(LockHeightFollow)
		db.Set(lockFollows[i], dbClient.Messages[i])
	}
	return lockFollows, nil
}

func RemoveLockHeightFollow(lockFollow *LockHeightFollow) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockFollow.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicMemoLockHeightFollow, [][]byte{lockFollow.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo follow", err)
	}
	return nil
}

func ListenLockHeightFollows(ctx context.Context, lockHashes [][]byte) (chan *LockHeightFollow, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockFollowChan = make(chan *LockHeightFollow)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(lockFollowChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoLockHeightFollow, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo follows by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockFollow = new(LockHeightFollow)
				db.Set(lockFollow, *msg)
				lockFollowChan <- lockFollow
			}
			cancelCtx.Cancel()
		}()
	}
	return lockFollowChan, nil
}
