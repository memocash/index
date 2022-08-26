package memo

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockHeightLike struct {
	LockHash   []byte
	Height     int64
	LikeTxHash []byte
	PostTxHash []byte
}

func (l LockHeightLike) GetUid() []byte {
	return jutil.CombineBytes(
		l.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(l.Height)),
		jutil.ByteReverse(l.LikeTxHash),
	)
}

func (l LockHeightLike) GetShard() uint {
	return client.GetByteShard(l.LockHash)
}

func (l LockHeightLike) GetTopic() string {
	return db.TopicMemoLockHeightLike
}

func (l LockHeightLike) Serialize() []byte {
	return l.PostTxHash
}

func (l *LockHeightLike) SetUid(uid []byte) {
	if len(uid) != memo.LockHashLength+memo.Int8Size+memo.TxHashLength {
		panic("invalid uid size for memo like")
	}
	l.LockHash = uid[:32]
	l.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	l.LikeTxHash = jutil.ByteReverse(uid[40:72])
}

func (l *LockHeightLike) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		panic("invalid data size for memo like")
	}
	l.PostTxHash = data
}

func RemoveLockHeightLike(lockLike *LockHeightLike) error {
	shardConfig := config.GetShardConfig(db.GetShard32(lockLike.GetShard()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.DeleteMessages(db.TopicMemoLockHeightLike, [][]byte{lockLike.GetUid()}); err != nil {
		return jerr.Get("error deleting item topic lock memo like", err)
	}
	return nil
}
