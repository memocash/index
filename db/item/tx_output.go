package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type TxOutput struct {
	TxHash   []byte
	Index    uint32
	Value    int64
	LockHash []byte
}

func (t TxOutput) GetUid() []byte {
	return GetTxOutputUid(t.TxHash, t.Index)
}

func (t TxOutput) GetShard() uint {
	return client.GetByteShard(t.TxHash)
}

func (t TxOutput) GetTopic() string {
	return TopicTxOutput
}

func (t TxOutput) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.GetInt64Data(t.Value),
		t.LockHash,
	)
}

func (t *TxOutput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	t.TxHash = jutil.ByteReverse(uid[:32])
	t.Index = jutil.GetUint32(uid[32:36])
}

func (t *TxOutput) Deserialize(data []byte) {
	if len(data) != 40 {
		return
	}
	t.Value = jutil.GetInt64(data[:8])
	t.LockHash = data[8:40]
}

func GetTxOutputUid(txHash []byte, index uint32) []byte {
	return GetTxHashIndexUid(txHash, index)
}

func GetTxOutputsByHashes(txHashes [][]byte) ([]*TxOutput, error) {
	var shardTxHashes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := uint32(GetShardByte(txHash))
		shardTxHashes[shard] = append(shardTxHashes[shard], jutil.ByteReverse(txHash))
	}
	var txOutputs []*TxOutput
	for shard, txHashes := range shardTxHashes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		err := db.GetByPrefixes(TopicTxOutput, txHashes)
		if err != nil {
			return nil, jerr.Get("error getting db message tx outputs", err)
		}
		for _, msg := range db.Messages {
			var txOutput = new(TxOutput)
			txOutput.SetUid(msg.Uid)
			txOutput.Deserialize(msg.Message)
			txOutputs = append(txOutputs, txOutput)
		}
	}
	return txOutputs, nil
}

func GetTxOutputs(outs []memo.Out) ([]*TxOutput, error) {
	var shardOutGroups = make(map[uint32][]memo.Out)
	for _, out := range outs {
		shard := GetShardByte32(out.TxHash)
		shardOutGroups[shard] = append(shardOutGroups[shard], out)
	}
	var txOutputs []*TxOutput
	for shard, outGroup := range shardOutGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		var uids = make([][]byte, len(outGroup))
		for i := range outGroup {
			uids[i] = GetTxOutputUid(outGroup[i].TxHash, outGroup[i].Index)
		}
		err := db.GetSpecific(TopicTxOutput, uids)
		if err != nil {
			return nil, jerr.Get("error getting db", err)
		}
		for i := range db.Messages {
			var txOutput = new(TxOutput)
			txOutput.SetUid(db.Messages[i].Uid)
			txOutput.Deserialize(db.Messages[i].Message)
			txOutputs = append(txOutputs, txOutput)
		}
	}
	return txOutputs, nil
}

func GetTxOutput(hash []byte, index uint32) (*TxOutput, error) {
	shard := GetShardByte32(hash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	uid := GetTxOutputUid(hash, index)
	if err := db.GetSingle(TopicTxOutput, uid); err != nil {
		return nil, jerr.Get("error getting db", err)
	}
	if len(db.Messages) != 1 {
		return nil, nil
	}
	var txOutput = new(TxOutput)
	txOutput.SetUid(db.Messages[0].Uid)
	txOutput.Deserialize(db.Messages[0].Message)
	return txOutput, nil
}
