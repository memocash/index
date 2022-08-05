package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockMemoName struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Name     string
}

func (n LockMemoName) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n LockMemoName) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n LockMemoName) GetTopic() string {
	return TopicLockMemoName
}

func (n LockMemoName) Serialize() []byte {
	return []byte(n.Name)
}

func (n *LockMemoName) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *LockMemoName) Deserialize(data []byte) {
	n.Name = string(data)
}

func GetLockMemoName(ctx context.Context, lockHash []byte) (*LockMemoName, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicLockMemoName,
		Prefixes: [][]byte{lockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db lock memo name by prefix", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error no lock memo names found", client.EntryNotFoundError)
	}
	var lockMemoName = new(LockMemoName)
	lockMemoName.SetUid(db.Messages[0].Uid)
	lockMemoName.Deserialize(db.Messages[0].Message)
	return lockMemoName, nil
}

func RemoveLockMemoName(lockMemoName *LockMemoName) error {
	shardConfig := config.GetShardConfig(GetShard32(lockMemoName.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicLockMemoName, [][]byte{lockMemoName.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo name", err)
	}
	return nil
}

func ListenLockMemoNames(ctx context.Context, lockHashes [][]byte) (chan *LockMemoName, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var lockMemoNameChan = make(chan *LockMemoName)
	cancelCtx := NewCancelContext(ctx, func() {
		close(lockMemoNameChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(cancelCtx.Context, TopicLockMemoName, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db lock memo names by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var lockMemoName = new(LockMemoName)
				Set(lockMemoName, *msg)
				lockMemoNameChan <- lockMemoName
			}
			cancelCtx.Cancel()
		}()
	}
	return lockMemoNameChan, nil
}
