package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"time"
)

type TxProcessed struct {
	TxHash    []byte
	Timestamp time.Time
}

func (s *TxProcessed) GetUid() []byte {
	return GetTxProcessedUid(s.TxHash, s.Timestamp)
}

func (s *TxProcessed) GetShardSource() uint {
	return client.GenShardSource(s.TxHash)
}

func (s *TxProcessed) GetTopic() string {
	return db.TopicChainTxProcessed
}

func (s *TxProcessed) Serialize() []byte {
	return nil
}

func (s *TxProcessed) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	s.TxHash = jutil.ByteReverse(uid[:32])
	s.Timestamp = jutil.GetByteTimeNanoBig(uid[32:40])
}

func (s *TxProcessed) Deserialize([]byte) {}

func GetTxProcessedUid(txHash []byte, timestamp time.Time) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash), jutil.GetTimeByteNanoBig(timestamp))
}

func GetTxProcessed(ctx context.Context, txHashes [][32]byte) ([]*TxProcessed, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicChainTxProcessed, db.ShardPrefixesTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client message chain tx processed; %w", err)
	}
	var txProcessedList []*TxProcessed
	for _, msg := range messages {
		var txProcessed = new(TxProcessed)
		db.Set(txProcessed, msg)
		txProcessedList = append(txProcessedList, txProcessed)
	}
	return txProcessedList, nil
}

func ListenTxProcessed(ctx context.Context, txHashes [][32]byte) (chan *TxProcessed, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range txHashes {
		shard := db.GetShardIdFromByte32(txHashes[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(txHashes[i][:]))
	}
	chanMessages, err := db.ListenPrefixes(ctx, db.TopicChainTxProcessed, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting listen prefixes for chain tx processed; %w", err)
	}
	var chanTxProcessed = make(chan *TxProcessed)
	go func() {
		for {
			msg, ok := <-chanMessages
			if !ok {
				close(chanTxProcessed)
				return
			}
			var txProcessed = new(TxProcessed)
			db.Set(txProcessed, *msg)
			chanTxProcessed <- txProcessed
		}
	}()
	return chanTxProcessed, nil
}
