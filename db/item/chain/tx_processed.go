package chain

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
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

func WaitForTxProcessed(ctx context.Context, txHash []byte) (*TxProcessed, error) {
	shardConfig := config.GetShardConfig(db.GetShardIdFromByte32(txHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Context:  ctx,
		Topic:    db.TopicChainTxProcessed,
		Prefixes: [][]byte{jutil.ByteReverse(txHash)},
		Wait:     true,
	}); err != nil {
		return nil, jerr.Get("error getting tx processed with wait db message", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error with tx processed wait, empty message", client.EntryNotFoundError)
	}
	var txProcessed = new(TxProcessed)
	db.Set(txProcessed, dbClient.Messages[0])
	return txProcessed, nil
}
