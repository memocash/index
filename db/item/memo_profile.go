package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoProfile struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Profile  string
}

func (n MemoProfile) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.GetInt64DataBig(n.Height),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n MemoProfile) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n MemoProfile) GetTopic() string {
	return TopicMemoProfile
}

func (n MemoProfile) Serialize() []byte {
	return []byte(n.Profile)
}

func (n *MemoProfile) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(uid[32:40])
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoProfile) Deserialize(data []byte) {
	n.Profile = string(data)
}

func GetMemoProfile(ctx context.Context, lockHash []byte) (*MemoProfile, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicMemoProfile,
		Prefixes: [][]byte{lockHash},
		Newest:   true,
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo profile by prefix", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error no memo profiles found", client.EntryNotFoundError)
	}
	var memoProfile = new(MemoProfile)
	memoProfile.SetUid(db.Messages[0].Uid)
	memoProfile.Deserialize(db.Messages[0].Message)
	return memoProfile, nil
}

func RemoveMemoProfile(memoProfile *MemoProfile) error {
	shardConfig := config.GetShardConfig(GetShard32(memoProfile.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicMemoProfile, [][]byte{memoProfile.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic memo profile", err)
	}
	return nil
}
