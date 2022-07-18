package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoProfilePic struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Pic      string
}

func (n MemoProfilePic) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n MemoProfilePic) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n MemoProfilePic) GetTopic() string {
	return TopicMemoProfilePic
}

func (n MemoProfilePic) Serialize() []byte {
	return []byte(n.Pic)
}

func (n *MemoProfilePic) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoProfilePic) Deserialize(data []byte) {
	n.Pic = string(data)
}

func GetMemoProfilePic(ctx context.Context, lockHash []byte) (*MemoProfilePic, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicMemoProfilePic,
		Prefixes: [][]byte{lockHash},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo profile pic by prefix", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error no memo profile pics found", client.EntryNotFoundError)
	}
	var memoProfilePic = new(MemoProfilePic)
	memoProfilePic.SetUid(db.Messages[0].Uid)
	memoProfilePic.Deserialize(db.Messages[0].Message)
	return memoProfilePic, nil
}

func RemoveMemoProfilePic(memoProfilePic *MemoProfilePic) error {
	shardConfig := config.GetShardConfig(GetShard32(memoProfilePic.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicMemoProfilePic, [][]byte{memoProfilePic.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic memo profile pic", err)
	}
	return nil
}

func ListenMemoProfilePics(ctx context.Context, lockHashes [][]byte) (chan *MemoProfilePic, error) {
	if len(lockHashes) == 0 {
		return nil, nil
	}
	var shardLockHashes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := client.GetByteShard32(lockHash)
		shardLockHashes[shard] = append(shardLockHashes[shard], lockHash)
	}
	shardConfigs := config.GetQueueShards()
	var memoProfilePicChan = make(chan *MemoProfilePic)
	for shard, lockHashPrefixes := range shardLockHashes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(ctx, TopicMemoProfilePic, lockHashPrefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo profile pic by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				if msg == nil {
					close(chanMessage)
					memoProfilePicChan <- nil
					break
				}
				var memoProfilePic = new(MemoProfilePic)
				memoProfilePic.SetUid(msg.Uid)
				memoProfilePic.Deserialize(msg.Message)
				memoProfilePicChan <- memoProfilePic
			}
		}()
	}
	return memoProfilePicChan, nil
}
