package item

import (
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
		jutil.GetInt64DataBig(n.Height),
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
	n.Height = jutil.GetInt64Big(uid[32:40])
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoProfilePic) Deserialize(data []byte) {
	n.Pic = string(data)
}

func GetMemoProfilePic(lockHash []byte) (*MemoProfilePic, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetByPrefix(TopicMemoProfilePic, lockHash); err != nil {
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
