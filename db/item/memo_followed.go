package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoFollowed struct {
	FollowLockHash []byte
	Height         int64
	TxHash         []byte
	LockHash       []byte
	Unfollow       bool
}

func (n MemoFollowed) GetUid() []byte {
	return jutil.CombineBytes(
		n.FollowLockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n MemoFollowed) GetShard() uint {
	return client.GetByteShard(n.FollowLockHash)
}

func (n MemoFollowed) GetTopic() string {
	return TopicMemoFollowed
}

func (n MemoFollowed) Serialize() []byte {
	var unfollow byte
	if n.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		n.LockHash,
	)
}

func (n *MemoFollowed) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.FollowLockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoFollowed) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+1 {
		return
	}
	n.Unfollow = data[0] == 1
	n.LockHash = data[1 : memo.TxHashLength+1]
}

func GetMemoFollowed(ctx context.Context, followLockHash []byte) (*MemoFollowed, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(followLockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicMemoFollowed,
		Prefixes: [][]byte{followLockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo followed by prefix", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error no memo followeds found", client.EntryNotFoundError)
	}
	var memoFollowed = new(MemoFollowed)
	Set(memoFollowed, db.Messages[0])
	return memoFollowed, nil
}

func GetMemoFolloweds(ctx context.Context, followLockHash []byte, start int64) ([]*MemoFollowed, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(followLockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	var startByte []byte
	if start != 0 {
		startByte = jutil.CombineBytes(followLockHash, jutil.ByteFlip(jutil.GetInt64DataBig(start)))
	} else {
		startByte = followLockHash
	}
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicMemoFollowed,
		Prefixes: [][]byte{followLockHash},
		Start:    startByte,
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo follow by prefix", err)
	}
	var memoFolloweds []*MemoFollowed
	for _, msg := range db.Messages {
		var memoFollow = new(MemoFollowed)
		Set(memoFollow, msg)
		memoFolloweds = append(memoFolloweds, memoFollow)
	}
	return memoFolloweds, nil
}

func RemoveMemoFollowed(memoFollow *MemoFollowed) error {
	shardConfig := config.GetShardConfig(GetShard32(memoFollow.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicMemoFollowed, [][]byte{memoFollow.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic memo followed", err)
	}
	return nil
}

func ListenMemoFolloweds(ctx context.Context, followLockHashes [][]byte) (chan *MemoFollowed, error) {
	var shardLockHashes = make(map[uint32][][]byte)
	for _, followLockHash := range followLockHashes {
		shard := client.GetByteShard32(followLockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], followLockHash)
	}
	shardConfigs := config.GetQueueShards()
	var memoFollowedChan = make(chan *MemoFollowed)
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		chanMessage, err := client.NewClient(shardConfig.GetHost()).Listen(ctx, TopicMemoFollowed, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo followeds by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var memoFollowed = new(MemoFollowed)
				Set(memoFollowed, *msg)
				memoFollowedChan <- memoFollowed
			}
			close(memoFollowedChan)
		}()
	}
	return memoFollowedChan, nil
}
