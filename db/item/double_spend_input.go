package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type DoubleSpendInput struct {
	TxHash []byte
	Index  uint32
}

func (i DoubleSpendInput) GetUid() []byte {
	return db.GetTxHashIndexUid(i.TxHash, i.Index)
}

func (i DoubleSpendInput) GetShard() uint {
	return client.GetByteShard(i.TxHash)
}

func (i DoubleSpendInput) GetTopic() string {
	return db.TopicDoubleSpendInput
}

func (i DoubleSpendInput) Serialize() []byte {
	return nil
}

func (i *DoubleSpendInput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	i.TxHash = jutil.ByteReverse(uid[:32])
	i.Index = jutil.GetUint32(uid[32:36])
}

func (i *DoubleSpendInput) Deserialize([]byte) {
	return
}

func GetDoubleSpendInputsByTxHashes(txHashes [][]byte) ([]*DoubleSpendInput, error) {
	var shardTxHashGroups = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardByte32(txHash)
		shardTxHashGroups[shard] = append(shardTxHashGroups[shard], txHash)
	}
	var doubleSpendInputs []*DoubleSpendInput
	for shard, outGroup := range shardTxHashGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		var prefixes = make([][]byte, len(outGroup))
		for i := range outGroup {
			prefixes[i] = jutil.ByteReverse(outGroup[i])
		}
		if err := dbClient.GetByPrefixes(db.TopicDoubleSpendInput, prefixes); err != nil {
			return nil, jerr.Get("error getting by prefixes for double spend inputs", err)
		}
		for i := range dbClient.Messages {
			var doubleSpendInput = new(DoubleSpendInput)
			db.Set(doubleSpendInput, dbClient.Messages[i])
			doubleSpendInputs = append(doubleSpendInputs, doubleSpendInput)
		}
	}
	return doubleSpendInputs, nil
}
