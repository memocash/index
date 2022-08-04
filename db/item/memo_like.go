package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoLike struct {
	LockHash   []byte
	Height     int64
	TxHash     []byte
	LikeTxHash []byte
}

func (n MemoLike) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash),
	)
}

func (n MemoLike) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n MemoLike) GetTopic() string {
	return TopicMemoLike
}

func (n MemoLike) Serialize() []byte {
	return n.LikeTxHash
}

func (n *MemoLike) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoLike) Deserialize(data []byte) {
	n.LikeTxHash = data
}

func RemoveMemoLike(memoLike *MemoLike) error {
	shardConfig := config.GetShardConfig(GetShard32(memoLike.GetShard()), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.DeleteMessages(TopicMemoLike, [][]byte{memoLike.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic memo like", err)
	}
	return nil
}
