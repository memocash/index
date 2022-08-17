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

type LockRoomFollow struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Unfollow bool
	Room     string
}

func (f LockRoomFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(f.Height)),
		jutil.ByteReverse(f.TxHash),
	)
}

func (f LockRoomFollow) GetShard() uint {
	return client.GetByteShard(f.LockHash)
}

func (f LockRoomFollow) GetTopic() string {
	return db.TopicLockMemoRoomFollow
}

func (f LockRoomFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		[]byte(f.Room),
	)
}

func (f *LockRoomFollow) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength*2+memo.Int8Size {
		return
	}
	f.LockHash = uid[:32]
	f.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	f.TxHash = jutil.ByteReverse(uid[40:72])
}

func (f *LockRoomFollow) Deserialize(data []byte) {
	if len(data) < 1 {
		return
	}
	f.Unfollow = data[0] == 1
	f.Room = string(data[1:])
}

func GetLockRoomFollows(ctx context.Context, lockHashes [][]byte) ([]*LockRoomFollow, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockFollows []*LockRoomFollow
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicLockMemoRoomFollow,
			Prefixes: prefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db memo lock room follow by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var lockFollow = new(LockRoomFollow)
			db.Set(lockFollow, msg)
			lockFollows = append(lockFollows, lockFollow)
		}
	}
	return lockFollows, nil
}

func ListenLockRoomFollows(ctx context.Context, lockHashes [][]byte) (chan *LockRoomFollow, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockRoomFollowChan = make(chan *LockRoomFollow)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(lockRoomFollowChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicLockMemoRoomFollow, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo lock room follow by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockRoomFollow = new(LockRoomFollow)
				db.Set(lockRoomFollow, *msg)
				lockRoomFollowChan <- lockRoomFollow
			}
			cancelCtx.Cancel()
		}()
	}
	return lockRoomFollowChan, nil
}
