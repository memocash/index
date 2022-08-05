package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoLikeTip struct {
	LikeTxHash []byte
	Tip        int64
}

func (n MemoLikeTip) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(n.LikeTxHash),
	)
}

func (n MemoLikeTip) GetShard() uint {
	return client.GetByteShard(n.LikeTxHash)
}

func (n MemoLikeTip) GetTopic() string {
	return TopicMemoLikeTip
}

func (n MemoLikeTip) Serialize() []byte {
	return jutil.GetInt64Data(n.Tip)
}

func (n *MemoLikeTip) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		panic("invalid uid size for memo like tip")
	}
	n.LikeTxHash = jutil.ByteReverse(uid[:32])
}

func (n *MemoLikeTip) Deserialize(data []byte) {
	if len(data) != memo.Int8Size {
		panic("invalid data size for memo like tip")
	}
	n.Tip = jutil.GetInt64(data)
}

func GetMemoLikeTips(likeTxHashes [][]byte) ([]*MemoLikeTip, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, likeTxHash := range likeTxHashes {
		shard := GetShardByte32(likeTxHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(likeTxHash))
	}
	var memoLikeTips []*MemoLikeTip
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetByPrefixes(TopicMemoLiked, prefixes); err != nil {
			return nil, jerr.Get("error getting client message memo likeds", err)
		}
		for _, msg := range db.Messages {
			var memoLikeTip = new(MemoLikeTip)
			memoLikeTip.SetUid(msg.Uid)
			memoLikeTips = append(memoLikeTips, memoLikeTip)
		}
	}
	return memoLikeTips, nil
}
