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
	return db.TopicLockMemoProfile
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
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicLockMemoProfile,
		Prefixes: [][]byte{lockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo profile by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no lock memo profiles found", client.EntryNotFoundError)
	}
	var lockMemoProfile = new(LockMemoProfile)
	db.Set(lockMemoProfile, dbClient.Messages[0])
	return lockMemoProfile, nil
}

func RemoveLockMemoProfile(lockMemoProfile *LockMemoProfile) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockMemoProfile.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicLockMemoProfile, [][]byte{lockMemoProfile.GetUid()}); err != nil {
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
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(lockMemoProfileChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicLockMemoProfile, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo profile by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockMemoProfile = new(LockMemoProfile)
				db.Set(lockMemoProfile, *msg)
				lockMemoProfileChan <- lockMemoProfile
			}
			cancelCtx.Cancel()
		}()
	}
	return lockMemoProfileChan, nil
}
