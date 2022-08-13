package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockMemoLike struct {
	LockHash   []byte
	Height     int64
	LikeTxHash []byte
	PostTxHash []byte
}

func (n LockMemoLike) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.LikeTxHash),
	)
}

func (n LockMemoLike) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n LockMemoLike) GetTopic() string {
	return db.TopicLockMemoLike
}

func (n LockMemoLike) Serialize() []byte {
	return n.PostTxHash
}

func (n *LockMemoLike) SetUid(uid []byte) {
	if len(uid) != memo.LockHashLength+memo.Int8Size+memo.TxHashLength {
		panic("invalid uid size for memo like")
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.LikeTxHash = jutil.ByteReverse(uid[40:72])
}

func (n *LockMemoLike) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		panic("invalid data size for memo like")
	}
	n.PostTxHash = data
}

func RemoveLockMemoLike(lockMemoLike *LockMemoLike) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockMemoLike.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicLockMemoLike, [][]byte{lockMemoLike.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo like", err)
	}
	return nil
}
