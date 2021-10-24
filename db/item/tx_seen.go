package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"time"
)

type TxSeen struct {
	TxHash    []byte
	Timestamp time.Time
}

func (s TxSeen) GetUid() []byte {
	return GetTxSeenUid(s.TxHash, s.Timestamp)
}

func (s TxSeen) GetShard() uint {
	return client.GetByteShard(s.TxHash)
}

func (s TxSeen) GetTopic() string {
	return TopicTxSeen
}

func (s TxSeen) Serialize() []byte {
	return nil
}

func (s *TxSeen) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	s.TxHash = jutil.ByteReverse(uid[:32])
	s.Timestamp = jutil.GetByteTime(uid[32:40])
}

func (s *TxSeen) Deserialize([]byte) {}

func GetTxSeenUid(txHash []byte, timestamp time.Time) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash), jutil.GetTimeByte(timestamp))
}
