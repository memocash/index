package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
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
			Start: jutil.ByteReverse(startTx),
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

func GetDoubleSpendsByOuts(outs []memo.Out) ([]*DoubleSpendOutput, error) {
	var shardOutGroups = make(map[uint32][]memo.Out)
	for _, out := range outs {
		shard := GetShardByte32(out.TxHash)
		shardOutGroups[shard] = append(shardOutGroups[shard], out)
	}
	var doubleSpendOutputs []*DoubleSpendOutput
	for shard, outGroup := range shardOutGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		var prefixes = make([][]byte, len(outGroup))
		for i := range outGroup {
			prefixes[i] = GetTxHashIndexUid(outGroup[i].TxHash, outGroup[i].Index)
		}
		if err := db.GetByPrefixes(TopicDoubleSpendOutput, prefixes); err != nil {
			return nil, jerr.Get("error getting by prefixes for double spend outputs", err)
		}
		for i := range db.Messages {
			var doubleSpendOutput = new(DoubleSpendOutput)
			doubleSpendOutput.SetUid(db.Messages[i].Uid)
			doubleSpendOutput.Deserialize(db.Messages[i].Message)
			doubleSpendOutputs = append(doubleSpendOutputs, doubleSpendOutput)
		}
	}
	return doubleSpendOutputs, nil
}
