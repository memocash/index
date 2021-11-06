package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
)

type TxInvalid struct {
	TxHash []byte
}

func (s TxInvalid) GetUid() []byte {
	return jutil.ByteReverse(s.TxHash)
}

func (s TxInvalid) GetShard() uint {
	return client.GetByteShard(s.TxHash)
}

func (s TxInvalid) GetTopic() string {
	return TopicTxInvalid
}

func (s TxInvalid) Serialize() []byte {
	return nil
}

func (s *TxInvalid) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	s.TxHash = jutil.ByteReverse(uid)
}

func (s *TxInvalid) Deserialize([]byte) {}

func GetTxInvalids(txHashes [][]byte) ([]*TxInvalid, error) {
	var shardTxHashGroups = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := GetShardByte32(txHash)
		shardTxHashGroups[shard] = append(shardTxHashGroups[shard], txHash)
	}
	var txInvalids []*TxInvalid
	for shard, outGroup := range shardTxHashGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		var uids = make([][]byte, len(outGroup))
		for i := range outGroup {
			uids[i] = jutil.ByteReverse(outGroup[i])
		}
		if err := db.GetSpecific(TopicTxInvalid, uids); err != nil {
			return nil, jerr.Get("error getting by uids for tx invalids", err)
		}
		for i := range db.Messages {
			var txInvalid = new(TxInvalid)
			txInvalid.SetUid(db.Messages[i].Uid)
			txInvalids = append(txInvalids, txInvalid)
		}
	}
	return txInvalids, nil
}

func RemoveTxInvalids(txHashes [][]byte) error {
	var shardUidsMap = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := uint32(GetShard(client.GetByteShard(txHash)))
		shardUidsMap[shard] = append(shardUidsMap[shard], jutil.ByteReverse(txHash))
	}
	for shard, shardUids := range shardUidsMap {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.DeleteMessages(TopicTxInvalid, shardUids); err != nil {
			return jerr.Get("error deleting topic tx invalids", err)
		}
	}
	return nil
}
