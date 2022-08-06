package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockMemoProfilePic struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Pic      string
}

func (n LockMemoProfilePic) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n LockMemoProfilePic) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n LockMemoProfilePic) GetTopic() string {
	return TopicLockMemoProfilePic
}

func (n LockMemoProfilePic) Serialize() []byte {
	return []byte(n.Pic)
}

func (n *LockMemoProfilePic) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *LockMemoProfilePic) Deserialize(data []byte) {
	n.Pic = string(data)
}

func GetLockMemoProfilePic(ctx context.Context, lockHash []byte) (*LockMemoProfilePic, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicLockMemoProfilePic,
		Prefixes: [][]byte{lockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo profile pic by prefix", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error no lock memo profile pics found", client.EntryNotFoundError)
	}
	var lockMemoProfilePic = new(LockMemoProfilePic)
	Set(lockMemoProfilePic, db.Messages[0])
	return lockMemoProfilePic, nil
}

func RemoveLockMemoProfilePic(lockMemoProfilePic *LockMemoProfilePic) error {
	shardConfig := config.GetShardConfig(GetShard32(lockMemoProfilePic.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicLockMemoProfilePic, [][]byte{lockMemoProfilePic.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo profile pic", err)
	}
	return nil
}

func ListenLockMemoProfilePics(ctx context.Context, lockHashes [][]byte) (chan *LockMemoProfilePic, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockMemoProfilePicChan = make(chan *LockMemoProfilePic)
	cancelCtx := NewCancelContext(ctx, func() {
		close(lockMemoProfilePicChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(cancelCtx.Context, TopicLockMemoProfilePic, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo profile pic by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockMemoProfilePic = new(LockMemoProfilePic)
				Set(lockMemoProfilePic, *msg)
				lockMemoProfilePicChan <- lockMemoProfilePic
			}
			cancelCtx.Cancel()
		}()
	}
	return lockMemoProfilePicChan, nil
}
