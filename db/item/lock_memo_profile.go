package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockMemoProfile struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Profile  string
}

func (n LockMemoProfile) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n LockMemoProfile) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n LockMemoProfile) GetTopic() string {
	return TopicLockMemoProfile
}

func (n LockMemoProfile) Serialize() []byte {
	return []byte(n.Profile)
}

func (n *LockMemoProfile) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *LockMemoProfile) Deserialize(data []byte) {
	n.Profile = string(data)
}

func GetLockMemoProfile(ctx context.Context, lockHash []byte) (*LockMemoProfile, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicLockMemoProfile,
		Prefixes: [][]byte{lockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo profile by prefix", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error no lock memo profiles found", client.EntryNotFoundError)
	}
	var lockMemoProfile = new(LockMemoProfile)
	Set(lockMemoProfile, db.Messages[0])
	return lockMemoProfile, nil
}

func RemoveLockMemoProfile(lockMemoProfile *LockMemoProfile) error {
	shardConfig := config.GetShardConfig(GetShard32(lockMemoProfile.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicLockMemoProfile, [][]byte{lockMemoProfile.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo profile", err)
	}
	return nil
}

func ListenLockMemoProfiles(ctx context.Context, lockHashes [][]byte) (chan *LockMemoProfile, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockMemoProfileChan = make(chan *LockMemoProfile)
	cancelCtx := NewCancelContext(ctx, func() {
		close(lockMemoProfileChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(cancelCtx.Context, TopicLockMemoProfile, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo profile by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockMemoProfile = new(LockMemoProfile)
				Set(lockMemoProfile, *msg)
				lockMemoProfileChan <- lockMemoProfile
			}
			cancelCtx.Cancel()
		}()
	}
	return lockMemoProfileChan, nil
}
