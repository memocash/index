package chain

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type TxOutput struct {
	TxHash     [32]byte
	Index      uint32
	Value      int64
	LockScript []byte
}

func (t *TxOutput) GetUid() []byte {
	return GetTxOutputUid(t.TxHash, t.Index)
}

func (t *TxOutput) GetShard() uint {
	return client.GetByteShard(t.TxHash[:])
}

func (t *TxOutput) GetTopic() string {
	return db.TopicChainTxOutput
}

func (t *TxOutput) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.GetInt64Data(t.Value),
		t.LockScript,
	)
}

func (t *TxOutput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	copy(t.TxHash[:], jutil.ByteReverse(uid[:32]))
	t.Index = jutil.GetUint32(uid[32:36])
}

func (t *TxOutput) Deserialize(data []byte) {
	if len(data) < 8 {
		return
	}
	t.Value = jutil.GetInt64(data[:8])
	t.LockScript = data[8:]
}

func GetTxOutputUid(txHash [32]byte, index uint32) []byte {
	return db.GetTxHashIndexUid(txHash[:], index)
}

func GetTxOutputsByHashes(txHashes [][]byte) ([]*TxOutput, error) {
	var shardTxHashes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := uint32(db.GetShardByte(txHash))
		shardTxHashes[shard] = append(shardTxHashes[shard], jutil.ByteReverse(txHash))
	}
	var txOutputs []*TxOutput
	for shard, txHashes := range shardTxHashes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		err := dbClient.GetByPrefixes(db.TopicChainTxOutput, txHashes)
		if err != nil {
			return nil, jerr.Get("error getting db message chain tx outputs", err)
		}
		for _, msg := range dbClient.Messages {
			var txOutput = new(TxOutput)
			db.Set(txOutput, msg)
			txOutputs = append(txOutputs, txOutput)
		}
	}
	return txOutputs, nil
}
