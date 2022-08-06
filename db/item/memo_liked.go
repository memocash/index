package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoLiked struct {
	PostTxHash []byte
	Height     int64
	LikeTxHash []byte
	LockHash   []byte
}

func (n MemoLiked) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(n.PostTxHash),
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.LikeTxHash),
	)
}

func (n MemoLiked) GetShard() uint {
	return client.GetByteShard(n.PostTxHash)
}

func (n MemoLiked) GetTopic() string {
	return TopicMemoLiked
}

func (n MemoLiked) Serialize() []byte {
	return n.LockHash
}

func (n *MemoLiked) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		panic("invalid uid size for memo liked")
	}
	n.PostTxHash = jutil.ByteReverse(uid[:32])
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	n.LikeTxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoLiked) Deserialize(data []byte) {
	if len(data) != memo.LockHashLength {
		panic("invalid data size for memo liked")
	}
	n.LockHash = data
}

func GetMemoLikeds(postTxHashes [][]byte) ([]*MemoLiked, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, postTxHash := range postTxHashes {
		shard := GetShardByte32(postTxHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHash))
	}
	var memoLikeds []*MemoLiked
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetByPrefixes(TopicMemoLiked, prefixes); err != nil {
			return nil, jerr.Get("error getting client message memo likeds", err)
		}
		for _, msg := range db.Messages {
			var memoLiked = new(MemoLiked)
			Set(memoLiked, msg)
			memoLikeds = append(memoLikeds, memoLiked)
		}
	}
	return memoLikeds, nil
}
