package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoName struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Name     string
}

func (n MemoName) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n MemoName) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n MemoName) GetTopic() string {
	return TopicMemoName
}

func (n MemoName) Serialize() []byte {
	return []byte(n.Name)
}

func (n *MemoName) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoName) Deserialize(data []byte) {
	n.Name = string(data)
}

func GetMemoName(ctx context.Context, lockHash []byte) (*MemoName, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicMemoName,
		Prefixes: [][]byte{lockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo name by prefix", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error no memo names found", client.EntryNotFoundError)
	}
	var memoName = new(MemoName)
	memoName.SetUid(db.Messages[0].Uid)
	memoName.Deserialize(db.Messages[0].Message)
	return memoName, nil
}

func RemoveMemoName(memoName *MemoName) error {
	shardConfig := config.GetShardConfig(GetShard32(memoName.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicMemoName, [][]byte{memoName.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic memo name", err)
	}
	return nil
}

func ListenMemoNames(ctx context.Context, lockHashes [][]byte) (chan *MemoName, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var memoNameChan = make(chan *MemoName)
	cancelCtx := NewCancelContext(ctx, func() {
		close(memoNameChan)
	})
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(cancelCtx.Context, TopicMemoName, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo names by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var memoName = new(MemoName)
				Set(memoName, *msg)
				memoNameChan <- memoName
			}
			cancelCtx.Cancel()
		}()
	}
	return memoNameChan, nil
}
