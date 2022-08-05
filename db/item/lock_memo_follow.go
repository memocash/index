package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockMemoFollow struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Unfollow bool
	Follow   []byte
}

func (n LockMemoFollow) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n LockMemoFollow) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n LockMemoFollow) GetTopic() string {
	return TopicLockMemoFollow
}

func (n LockMemoFollow) Serialize() []byte {
	var unfollow byte
	if n.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		n.Follow,
	)
}

func (n *LockMemoFollow) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *LockMemoFollow) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+1 {
		return
	}
	n.Unfollow = data[0] == 1
	n.Follow = data[1 : memo.TxHashLength+1]
}

func GetLockMemoFollow(ctx context.Context, lockHashes [][]byte) ([]*LockMemoFollow, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockMemoFollows []*LockMemoFollow
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetWOpts(client.Opts{
			Topic:    TopicLockMemoFollow,
			Prefixes: lockHashPrefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db lock memo follow by prefix", err)
		}
		for _, msg := range db.Messages {
			var lockMemoFollow = new(LockMemoFollow)
			Set(lockMemoFollow, msg)
			lockMemoFollows = append(lockMemoFollows, lockMemoFollow)
		}
	}
	return lockMemoFollows, nil
}

func GetLockMemoFollows(ctx context.Context, lockHash []byte, start int64) ([]*LockMemoFollow, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	var startByte []byte
	if start != 0 {
		startByte = jutil.CombineBytes(lockHash, jutil.ByteFlip(jutil.GetInt64DataBig(start)))
	} else {
		startByte = lockHash
	}
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicLockMemoFollow,
		Prefixes: [][]byte{lockHash},
		Start:    startByte,
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo follow by prefix", err)
	}
	var lockMemoFollows []*LockMemoFollow
	for _, msg := range db.Messages {
		var lockMemoFollow = new(LockMemoFollow)
		Set(lockMemoFollow, msg)
		lockMemoFollows = append(lockMemoFollows, lockMemoFollow)
	}
	return lockMemoFollows, nil
}

func RemoveLockMemoFollow(lockMemoFollow *LockMemoFollow) error {
	shardConfig := config.GetShardConfig(GetShard32(lockMemoFollow.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicLockMemoFollow, [][]byte{lockMemoFollow.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo follow", err)
	}
	return nil
}

func ListenLockMemoFollows(ctx context.Context, lockHashes [][]byte) (chan *LockMemoFollow, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockMemoFollowChan = make(chan *LockMemoFollow)
	cancelCtx := NewCancelContext(ctx, func() {
		close(lockMemoFollowChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(cancelCtx.Context, TopicLockMemoFollow, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo follows by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockMemoFollow = new(LockMemoFollow)
				Set(lockMemoFollow, *msg)
				lockMemoFollowChan <- lockMemoFollow
			}
			cancelCtx.Cancel()
		}()
	}
	return lockMemoFollowChan, nil
}
