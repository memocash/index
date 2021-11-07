package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
)

type TxSuspect struct {
	TxHash []byte
}

func (s TxSuspect) GetUid() []byte {
	return jutil.ByteReverse(s.TxHash)
}

func (s TxSuspect) GetShard() uint {
	return client.GetByteShard(s.TxHash)
}

func (s TxSuspect) GetTopic() string {
	return TopicTxSuspect
}

func (s TxSuspect) Serialize() []byte {
	return nil
}

func (s *TxSuspect) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	s.TxHash = jutil.ByteReverse(uid)
}

func (s *TxSuspect) Deserialize([]byte) {}

func GetTxSuspects(txHashes [][]byte) ([]*TxSuspect, error) {
	var shardTxHashGroups = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := GetShardByte32(txHash)
		shardTxHashGroups[shard] = append(shardTxHashGroups[shard], txHash)
	}
	var txSuspects []*TxSuspect
	for shard, outGroup := range shardTxHashGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		var uids = make([][]byte, len(outGroup))
		for i := range outGroup {
			uids[i] = jutil.ByteReverse(outGroup[i])
		}
		if err := db.GetSpecific(TopicTxSuspect, uids); err != nil {
			return nil, jerr.Get("error getting by uids for tx suspects", err)
		}
		for i := range db.Messages {
			var txSuspect = new(TxSuspect)
			txSuspect.SetUid(db.Messages[i].Uid)
			txSuspects = append(txSuspects, txSuspect)
		}
	}
	return txSuspects, nil
}
