package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
)

type DoubleSpendOutput struct {
	TxHash []byte
	Index  uint32
}

func (o DoubleSpendOutput) GetUid() []byte {
	return GetTxHashIndexUid(o.TxHash, o.Index)
}

func (o DoubleSpendOutput) GetShard() uint {
	return client.GetByteShard(o.TxHash)
}

func (o DoubleSpendOutput) GetTopic() string {
	return TopicDoubleSpendOutput
}

func (o DoubleSpendOutput) Serialize() []byte {
	return nil
}

func (o *DoubleSpendOutput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	o.TxHash = jutil.ByteReverse(uid[:32])
	o.Index = jutil.GetUint32(uid[32:36])
}

func (o *DoubleSpendOutput) Deserialize([]byte) {
	return
}

// GetDoubleSpendOutputs begins on shard 0 if no start tx specified.
// If the limit is not reached it will move onto the next shard.
// If the start tx is specified, results will begin with the shard of the start tx.
func GetDoubleSpendOutputs(startTx []byte, limit uint32) ([]*DoubleSpendOutput, error) {
	var startShard uint32
	if len(startTx) > 0 {
		startShard = client.GetByteShard32(startTx)
	}
	configQueueShards := config.GetQueueShards()
	startShardConfig := config.GetShardConfig(startShard, configQueueShards)
	if limit == 0 {
		limit = client.LargeLimit
	}
	var doubleSpendOutputs []*DoubleSpendOutput
	for shard := startShardConfig.Min; shard < startShardConfig.Total; shard++ {
		shardConfig := config.GetShardConfig(shard, configQueueShards)
		db := client.NewClient(shardConfig.GetHost())
		err := db.GetWOpts(client.Opts{
			Topic: TopicDoubleSpendOutput,
			Start: startTx,
			Max:   limit,
		})
		if err != nil {
			return nil, jerr.Get("error getting db message for double spend outputs", err)
		}
		for _, msg := range db.Messages {
			doubleSpendOutput := new(DoubleSpendOutput)
			doubleSpendOutput.SetUid(msg.Uid)
			doubleSpendOutputs = append(doubleSpendOutputs, doubleSpendOutput)
		}
		limit -= uint32(len(db.Messages))
		if limit <= 0 {
			break
		}
	}
	return doubleSpendOutputs, nil
}
