package item

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
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
	if bytes.Equal(uid[36:40], []byte{0x0, 0x0, 0x0, 0x0}) {
		s.Timestamp = jutil.GetByteTime(uid[32:40])
	} else {
		s.Timestamp = jutil.GetByteTimeNano(uid[32:40])
	}
}

func (s *TxSeen) Deserialize([]byte) {}

func GetTxSeenUid(txHash []byte, timestamp time.Time) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash), jutil.GetTimeByteNano(timestamp))
}

func GetTxSeens(txHashes [][]byte) ([]*TxSeen, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := GetShardByte32(txHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(txHash))
	}
	var txSeens []*TxSeen
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetByPrefixes(TopicTxSeen, prefixes); err != nil {
			return nil, jerr.Get("error getting client message tx seens", err)
		}
		for _, msg := range db.Messages {
			var txSeen = new(TxSeen)
			txSeen.SetUid(msg.Uid)
			txSeens = append(txSeens, txSeen)
		}
	}
	return txSeens, nil
}
