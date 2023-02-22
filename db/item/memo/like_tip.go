package memo

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LikeTip struct {
	LikeTxHash [32]byte
	Tip        int64
}

func (t *LikeTip) GetTopic() string {
	return db.TopicMemoLikeTip
}

func (t *LikeTip) GetShard() uint {
	return client.GetByteShard(t.LikeTxHash[:])
}

func (t *LikeTip) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.LikeTxHash[:]),
	)
}

func (t *LikeTip) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		panic("invalid uid size for memo like tip")
	}
	copy(t.LikeTxHash[:], jutil.ByteReverse(uid[:32]))
}

func (t *LikeTip) Serialize() []byte {
	return jutil.GetInt64Data(t.Tip)
}

func (t *LikeTip) Deserialize(data []byte) {
	if len(data) != memo.Int8Size {
		panic("invalid data size for memo like tip")
	}
	t.Tip = jutil.GetInt64(data)
}

func GetLikeTips(likeTxHashes [][32]byte) ([]*LikeTip, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range likeTxHashes {
		shard := db.GetShardByte32(likeTxHashes[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(likeTxHashes[i][:]))
	}
	var likeTips []*LikeTip
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetByPrefixes(db.TopicMemoLikeTip, prefixes); err != nil {
			return nil, jerr.Get("error getting client message memo like tips", err)
		}
		for _, msg := range dbClient.Messages {
			var likeTip = new(LikeTip)
			db.Set(likeTip, msg)
			likeTips = append(likeTips, likeTip)
		}
	}
	return likeTips, nil
}
