package chain

import (
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
