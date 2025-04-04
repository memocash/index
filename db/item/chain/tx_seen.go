package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"time"
)

type TxSeen struct {
	TxHash    [32]byte
	Timestamp time.Time
}

func (s *TxSeen) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(s.TxHash[:]),
		jutil.GetTimeByteNanoBig(s.Timestamp),
	)
}

func (s *TxSeen) GetShardSource() uint {
	return client.GenShardSource(s.TxHash[:])
}

func (s *TxSeen) GetTopic() string {
	return db.TopicChainTxSeen
}

func (s *TxSeen) Serialize() []byte {
	return nil
}

func (s *TxSeen) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	copy(s.TxHash[:], jutil.ByteReverse(uid[:32]))
	s.Timestamp = jutil.GetByteTimeNanoBig(uid[32:40])
}

func (s *TxSeen) Deserialize([]byte) {}

func GetTxSeens(ctx context.Context, txHashes [][32]byte) ([]*TxSeen, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicChainTxSeen, db.ShardPrefixesTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client message chain tx seen; %w", err)
	}
	var txSeens []*TxSeen
	for _, msg := range messages {
		var txSeen = new(TxSeen)
		db.Set(txSeen, msg)
		txSeens = append(txSeens, txSeen)
	}
	return txSeens, nil
}
