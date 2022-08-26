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

type LockHeightProfile struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Profile  string
}

func (p LockHeightProfile) GetUid() []byte {
	return jutil.CombineBytes(
		p.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash),
	)
}

func (p LockHeightProfile) GetShard() uint {
	return client.GetByteShard(p.LockHash)
}

func (p LockHeightProfile) GetTopic() string {
	return db.TopicLockMemoProfile
}

func (p LockHeightProfile) Serialize() []byte {
	return []byte(p.Profile)
}

func (p *LockHeightProfile) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	p.LockHash = uid[:32]
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	p.TxHash = jutil.ByteReverse(uid[40:72])
}

func (p *LockHeightProfile) Deserialize(data []byte) {
	p.Profile = string(data)
}

func GetLockHeightProfile(ctx context.Context, lockHash []byte) (*LockHeightProfile, error) {
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
	var lockProfile = new(LockHeightProfile)
	db.Set(lockProfile, dbClient.Messages[0])
	return lockProfile, nil
}

func RemoveLockHeightProfile(lockProfile *LockHeightProfile) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockProfile.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicLockMemoProfile, [][]byte{lockProfile.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo profile", err)
	}
	return nil
}

func ListenLockHeightProfiles(ctx context.Context, lockHashes [][]byte) (chan *LockHeightProfile, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockProfileChan = make(chan *LockHeightProfile)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(lockProfileChan)
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
				var lockProfile = new(LockHeightProfile)
				db.Set(lockProfile, *msg)
				lockProfileChan <- lockProfile
			}
			cancelCtx.Cancel()
		}()
	}
	return lockProfileChan, nil
}
