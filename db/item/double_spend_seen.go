package item

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"sort"
	"time"
)

type DoubleSpendSeen struct {
	Timestamp time.Time
	TxHash    []byte
	Index     uint32
}

func (s DoubleSpendSeen) GetUid() []byte {
	return GetDoubleSpendSeenUid(s.Timestamp, s.TxHash, s.Index)
}

func (s DoubleSpendSeen) GetShard() uint {
	return client.GetByteShard(s.TxHash)
}

func (s DoubleSpendSeen) GetTopic() string {
	return TopicDoubleSpendSeen
}

func (s DoubleSpendSeen) Serialize() []byte {
	return nil
}

func (s *DoubleSpendSeen) SetUid(uid []byte) {
	if len(uid) != 44 {
		return
	}
	var ts = uid[:8]
	if bytes.Equal(ts, []byte{0x0, 0x0, 0x0, 0x0}) {
		s.Timestamp = jutil.GetByteTime(ts)
	} else {
		s.Timestamp = jutil.GetByteTimeNano(ts)
	}
	s.TxHash = jutil.ByteReverse(uid[8:40])
	s.Index = jutil.GetUint32(uid[40:44])
}

func (s *DoubleSpendSeen) Deserialize([]byte) {}

func GetDoubleSpendSeenUid(timestamp time.Time, txHash []byte, index uint32) []byte {
	return jutil.CombineBytes(jutil.GetTimeByteNano(timestamp), jutil.ByteReverse(txHash), jutil.GetUint32Data(index))
}

func GetDoubleSpendSeensAllLimit(startTime time.Time, limit uint32, newest bool) ([]*DoubleSpendSeen, error) {
	var doubleSpendSeens []*DoubleSpendSeen
	shardConfigs := config.GetQueueShards()
	shardLimit := limit / uint32(len(shardConfigs))
	for _, shardConfig := range shardConfigs {
		dbClient := client.NewClient(shardConfig.GetHost())
		var start []byte
		if !startTime.IsZero() {
			start = jutil.GetTimeByteNano(startTime)
		}
		err := dbClient.GetWOpts(client.Opts{
			Topic:  TopicDoubleSpendSeen,
			Start:  start,
			Max:    shardLimit,
			Newest: newest,
		})
		if err != nil {
			return nil, jerr.Get("error getting double spend seens from queue client all", err)
		}
		for i := range dbClient.Messages {
			var doubleSpendSeen = new(DoubleSpendSeen)
			doubleSpendSeen.SetUid(dbClient.Messages[i].Uid)
			doubleSpendSeen.Deserialize(dbClient.Messages[i].Message)
			doubleSpendSeens = append(doubleSpendSeens, doubleSpendSeen)
		}
	}
	sort.Slice(doubleSpendSeens, func(i, j int) bool {
		if newest {
			return doubleSpendSeens[i].Timestamp.After(doubleSpendSeens[j].Timestamp)
		} else {
			return doubleSpendSeens[i].Timestamp.Before(doubleSpendSeens[j].Timestamp)
		}
	})
	return doubleSpendSeens, nil
}

func GetDoubleSpendSeensByTxHashesScanAll(txHashes [][]byte) ([]*DoubleSpendSeen, error) {
	var shardTxHashGroups = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := GetShardByte32(txHash)
		shardTxHashGroups[shard] = append(shardTxHashGroups[shard], txHash)
	}
	var doubleSpendSeens []*DoubleSpendSeen
	for shard, txHashGroup := range shardTxHashGroups {
		sort.Slice(txHashGroup, func(i, j int) bool {
			return bytes.Compare(txHashGroup[i], txHashGroup[j]) == -1
		})
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		var txHashId int
		var startTxHash = txHashGroup[txHashId]
		for {
			if err := db.Get(TopicDoubleSpendSeen, startTxHash, false); err != nil {
				return nil, jerr.Get("error getting by double spend seens for scan all", err)
			}
			for i := range db.Messages {
				var doubleSpendSeen = new(DoubleSpendSeen)
				doubleSpendSeen.SetUid(db.Messages[i].Uid)
				for ; txHashId < len(txHashGroup) && bytes.Compare(doubleSpendSeen.TxHash, txHashGroup[txHashId]) != -1; txHashId++ {
					startTxHash = txHashGroup[txHashId]
				}
				if bytes.Equal(doubleSpendSeen.TxHash, startTxHash) {
					doubleSpendSeens = append(doubleSpendSeens, doubleSpendSeen)
				}
			}
			if len(db.Messages) < client.DefaultLimit {
				break
			}
		}
	}
	return doubleSpendSeens, nil
}
