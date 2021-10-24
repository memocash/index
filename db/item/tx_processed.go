package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
	"time"
)

type TxProcessed struct {
	TxHash    []byte
	Timestamp time.Time
}

func (s TxProcessed) GetUid() []byte {
	return GetTxProcessedUid(s.TxHash, s.Timestamp)
}

func (s TxProcessed) GetShard() uint {
	return client.GetByteShard(s.TxHash)
}

func (s TxProcessed) GetTopic() string {
	return TopicTxProcessed
}

func (s TxProcessed) Serialize() []byte {
	return nil
}

func (s *TxProcessed) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	s.TxHash = jutil.ByteReverse(uid[:32])
	s.Timestamp = jutil.GetByteTime(uid[32:40])
}

func (s *TxProcessed) Deserialize([]byte) {}

func GetTxProcessedUid(txHash []byte, timestamp time.Time) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash), jutil.GetTimeByte(timestamp))
}

func WaitForTxProcessed(ctx context.Context, txHash []byte) (*TxProcessed, error) {
	shardConfig := config.GetShardConfig(GetShardByte32(txHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	err := db.GetWOpts(client.Opts{
		Context:  ctx,
		Topic:    TopicTxProcessed,
		Prefixes: [][]byte{jutil.ByteReverse(txHash)},
		Wait:     true,
	})
	if err != nil {
		return nil, jerr.Get("error getting tx processed with wait db message", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error with tx processed wait, empty message", client.EntryNotFoundError)
	}

	var txProcessed = new(TxProcessed)
	txProcessed.SetUid(db.Messages[0].Uid)
	txProcessed.Deserialize(db.Messages[0].Message)
	return txProcessed, nil
}
