package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type DoubleSpendOutput struct {
	TxHash []byte
	Index  uint32
}

func (o DoubleSpendOutput) GetUid() []byte {
	return db.GetTxHashIndexUid(o.TxHash, o.Index)
}

func (o DoubleSpendOutput) GetShard() uint {
	return client.GetByteShard(o.TxHash)
}

func (o DoubleSpendOutput) GetTopic() string {
	return db.TopicDoubleSpendOutput
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
func GetDoubleSpendOutputs(start *DoubleSpendOutput, limit uint32) ([]*DoubleSpendOutput, error) {
	var startShard uint32
	var startUid []byte
	if start != nil {
		startShard = client.GetByteShard32(start.TxHash)
		startUid = start.GetUid()
	}
	configQueueShards := config.GetQueueShards()
	startShardConfig := config.GetShardConfig(startShard, configQueueShards)
	if limit == 0 {
		limit = client.LargeLimit
	}
	var doubleSpendOutputs []*DoubleSpendOutput
	for shard := startShardConfig.Shard; shard < startShardConfig.Total; shard++ {
		if shard > startShardConfig.Shard {
			startUid = nil
		}
		shardConfig := config.GetShardConfig(shard, configQueueShards)
		dbClient := client.NewClient(shardConfig.GetHost())
		err := dbClient.GetWOpts(client.Opts{
			Topic: db.TopicDoubleSpendOutput,
			Start: startUid,
			Max:   limit,
		})
		if err != nil {
			return nil, jerr.Get("error getting db message for double spend outputs", err)
		}
		for _, msg := range dbClient.Messages {
			doubleSpendOutput := new(DoubleSpendOutput)
			db.Set(doubleSpendOutput, msg)
			doubleSpendOutputs = append(doubleSpendOutputs, doubleSpendOutput)
		}
		limit -= uint32(len(dbClient.Messages))
		if limit <= 0 {
			break
		}
	}
	return doubleSpendOutputs, nil
}

func GetDoubleSpendsByOuts(outs []memo.Out) ([]*DoubleSpendOutput, error) {
	var shardOutGroups = make(map[uint32][]memo.Out)
	for _, out := range outs {
		shard := db.GetShardByte32(out.TxHash)
		shardOutGroups[shard] = append(shardOutGroups[shard], out)
	}
	var doubleSpendOutputs []*DoubleSpendOutput
	for shard, outGroup := range shardOutGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		var prefixes = make([][]byte, len(outGroup))
		for i := range outGroup {
			prefixes[i] = db.GetTxHashIndexUid(outGroup[i].TxHash, outGroup[i].Index)
		}
		if err := dbClient.GetByPrefixes(db.TopicDoubleSpendOutput, prefixes); err != nil {
			return nil, jerr.Get("error getting by prefixes for double spend outputs", err)
		}
		for i := range dbClient.Messages {
			var doubleSpendOutput = new(DoubleSpendOutput)
			db.Set(doubleSpendOutput, dbClient.Messages[i])
			doubleSpendOutputs = append(doubleSpendOutputs, doubleSpendOutput)
		}
	}
	return doubleSpendOutputs, nil
}
