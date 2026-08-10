package slp

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

func GetTopics() []db.Object {
	return []db.Object{
		&Genesis{},
		&Mint{},
		&Send{},
		&Output{},
		&Baton{},
		&Validity{},
	}
}

// GetTopicTxHashes returns a page of tx hashes from one of the SLP
// transcription topics (genesis/mint/send — uid is the reversed tx hash) on a
// single shard, starting at startUid inclusive. A page shorter than
// client.HugeLimit means the shard's topic is exhausted.
func GetTopicTxHashes(ctx context.Context, topic string, shard uint32, startUid []byte) ([][32]byte, error) {
	dbClient := db.GetShardClient(shard)
	if err := dbClient.GetAll(ctx, topic, startUid, client.OptionHugeLimit()); err != nil {
		return nil, fmt.Errorf("error getting slp topic tx hashes: %s; %w", topic, err)
	}
	var txHashes = make([][32]byte, len(dbClient.Messages))
	for i := range dbClient.Messages {
		if len(dbClient.Messages[i].Uid) != memo.TxHashLength {
			return nil, fmt.Errorf("error unexpected uid length %d for slp topic: %s",
				len(dbClient.Messages[i].Uid), topic)
		}
		copy(txHashes[i][:], jutil.ByteReverse(dbClient.Messages[i].Uid))
	}
	return txHashes, nil
}
