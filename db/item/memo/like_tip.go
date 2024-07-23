package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LikeTip struct {
	LikeTxHash [32]byte
	Tip        int64
}

func (t *LikeTip) GetTopic() string {
	return db.TopicMemoLikeTip
}

func (t *LikeTip) GetShardSource() uint {
	return client.GenShardSource(t.LikeTxHash[:])
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

func GetLikeTips(ctx context.Context, likeTxHashes [][32]byte) ([]*LikeTip, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicMemoLikeTip, db.ShardPrefixesTxHashes(likeTxHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client messages memo like tips; %w", err)
	}
	var likeTips = make([]*LikeTip, len(messages))
	for i := range messages {
		likeTips[i] = new(LikeTip)
		db.Set(likeTips[i], messages[i])
	}
	return likeTips, nil
}
