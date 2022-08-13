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

type LockProfilePic struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Pic      string
}

func (p LockProfilePic) GetUid() []byte {
	return jutil.CombineBytes(
		p.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash),
	)
}

func (p LockProfilePic) GetShard() uint {
	return client.GetByteShard(p.LockHash)
}

func (p LockProfilePic) GetTopic() string {
	return db.TopicLockMemoProfilePic
}

func (p LockProfilePic) Serialize() []byte {
	return []byte(p.Pic)
}

func (p *LockProfilePic) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	p.LockHash = uid[:32]
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	p.TxHash = jutil.ByteReverse(uid[40:72])
}

func (p *LockProfilePic) Deserialize(data []byte) {
	p.Pic = string(data)
}

func GetLockProfilePic(ctx context.Context, lockHash []byte) (*LockProfilePic, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicLockMemoProfilePic,
		Prefixes: [][]byte{lockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo profile pic by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no lock memo profile pics found", client.EntryNotFoundError)
	}
	var lockProfilePic = new(LockProfilePic)
	db.Set(lockProfilePic, dbClient.Messages[0])
	return lockProfilePic, nil
}

func RemoveLockProfilePic(lockProfilePic *LockProfilePic) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockProfilePic.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicLockMemoProfilePic, [][]byte{lockProfilePic.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo profile pic", err)
	}
	return nil
}

func ListenLockProfilePics(ctx context.Context, lockHashes [][]byte) (chan *LockProfilePic, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockProfilePicChan = make(chan *LockProfilePic)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(lockProfilePicChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicLockMemoProfilePic, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo profile pic by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockProfilePic = new(LockProfilePic)
				db.Set(lockProfilePic, *msg)
				lockProfilePicChan <- lockProfilePic
			}
			cancelCtx.Cancel()
		}()
	}
	return lockProfilePicChan, nil
}
