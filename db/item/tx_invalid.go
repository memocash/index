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
