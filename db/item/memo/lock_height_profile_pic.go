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

type LockHeightProfilePic struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Pic      string
}

func (p LockHeightProfilePic) GetUid() []byte {
	return jutil.CombineBytes(
		p.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash),
	)
}

func (p LockHeightProfilePic) GetShard() uint {
	return client.GetByteShard(p.LockHash)
}

func (p LockHeightProfilePic) GetTopic() string {
	return db.TopicMemoLockHeightProfilePic
}

func (p LockHeightProfilePic) Serialize() []byte {
	return []byte(p.Pic)
}

func (p *LockHeightProfilePic) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	p.LockHash = uid[:32]
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	p.TxHash = jutil.ByteReverse(uid[40:72])
}

func (p *LockHeightProfilePic) Deserialize(data []byte) {
	p.Pic = string(data)
}

func GetLockHeightProfilePic(ctx context.Context, lockHash []byte) (*LockHeightProfilePic, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoLockHeightProfilePic,
		Prefixes: [][]byte{lockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo profile pic by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no lock memo profile pics found", client.EntryNotFoundError)
	}
	var lockProfilePic = new(LockHeightProfilePic)
	db.Set(lockProfilePic, dbClient.Messages[0])
	return lockProfilePic, nil
}

func RemoveLockHeightProfilePic(lockProfilePic *LockHeightProfilePic) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockProfilePic.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicMemoLockHeightProfilePic, [][]byte{lockProfilePic.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo profile pic", err)
	}
	return nil
}

func ListenLockHeightProfilePics(ctx context.Context, lockHashes [][]byte) (chan *LockHeightProfilePic, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockProfilePicChan = make(chan *LockHeightProfilePic)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(lockProfilePicChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoLockHeightProfilePic, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo profile pic by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockProfilePic = new(LockHeightProfilePic)
				db.Set(lockProfilePic, *msg)
				lockProfilePicChan <- lockProfilePic
			}
			cancelCtx.Cancel()
		}()
	}
	return lockProfilePicChan, nil
}
