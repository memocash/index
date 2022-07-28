package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoFollow struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Unfollow bool
	Follow   []byte
}

func (n MemoFollow) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n MemoFollow) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n MemoFollow) GetTopic() string {
	return TopicMemoFollow
}

func (n MemoFollow) Serialize() []byte {
	var unfollow byte
	if n.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		n.Follow,
	)
}

func (n *MemoFollow) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoFollow) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+1 {
		return
	}
	n.Unfollow = data[0] == 1
	n.Follow = data[1 : memo.TxHashLength+1]
}

func GetMemoFollow(ctx context.Context, lockHashes [][]byte) ([]*MemoFollow, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var memoFollows []*MemoFollow
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetWOpts(client.Opts{
			Topic:    TopicMemoFollow,
			Prefixes: lockHashPrefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db memo follow by prefix", err)
		}
		for _, msg := range db.Messages {
			var memoFollow = new(MemoFollow)
			Set(memoFollow, msg)
			memoFollows = append(memoFollows, memoFollow)
		}
	}
	return memoFollows, nil
}

func GetMemoFollows(ctx context.Context, lockHash []byte, start int64) ([]*MemoFollow, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	var startByte []byte
	if start != 0 {
		startByte = jutil.CombineBytes(lockHash, jutil.ByteFlip(jutil.GetInt64DataBig(start)))
	} else {
		startByte = lockHash
	}
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicMemoFollow,
		Prefixes: [][]byte{lockHash},
		Start:    startByte,
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo follow by prefix", err)
	}
	var memoFollows []*MemoFollow
	for _, msg := range db.Messages {
		var memoFollow = new(MemoFollow)
		Set(memoFollow, msg)
		memoFollows = append(memoFollows, memoFollow)
	}
	return memoFollows, nil
}

func RemoveMemoFollow(memoFollow *MemoFollow) error {
	shardConfig := config.GetShardConfig(GetShard32(memoFollow.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicMemoFollow, [][]byte{memoFollow.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic memo follow", err)
	}
	return nil
}

func ListenMemoFollows(ctx context.Context, lockHashes [][]byte) (chan *MemoFollow, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var memoFollowChan = make(chan *MemoFollow)
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(ctx, TopicMemoFollow, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo follows by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var memoFollow = new(MemoFollow)
				memoFollow.SetUid(msg.Uid)
				memoFollow.Deserialize(msg.Message)
				memoFollowChan <- memoFollow
			}
			close(memoFollowChan)
		}()
	}
	return memoFollowChan, nil
}
