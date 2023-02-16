package chain

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"sort"
	"time"
)

type TxSeen struct {
	TxHash    [32]byte
	Timestamp time.Time
}

func (s *TxSeen) GetUid() []byte {
	return GetTxSeenUid(s.TxHash[:], s.Timestamp)
}

func (s *TxSeen) GetShard() uint {
	return client.GetByteShard(s.TxHash[:])
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

func GetTxSeens(txHashes [][32]byte) ([]*TxSeen, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardByte32(txHash[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(txHash[:]))
	}
	var txSeens []*TxSeen
	for shard, prefixes := range shardPrefixes {
		sort.Slice(prefixes, func(i, j int) bool {
			return jutil.ByteLT(prefixes[i], prefixes[j])
		})
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetByPrefixes(db.TopicChainTxSeen, prefixes); err != nil {
			return nil, jerr.Get("error getting client message tx seens", err)
		}
		for _, msg := range dbClient.Messages {
			var txSeen = new(TxSeen)
			db.Set(txSeen, msg)
			txSeens = append(txSeens, txSeen)
		}
	}
	return txSeens, nil
}
