package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type TxInput struct {
	TxHash    []byte
	Index     uint32
	PrevHash  []byte
	PrevIndex uint32
}

func (t TxInput) GetUid() []byte {
	return GetTxInputUid(t.TxHash, t.Index)
}

func (t TxInput) GetShard() uint {
	return client.GetByteShard(t.TxHash)
}

func (t TxInput) GetTopic() string {
	return TopicTxInput
}

func (t TxInput) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.PrevHash),
		jutil.GetUint32Data(t.PrevIndex),
	)
}

func (t *TxInput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	t.TxHash = jutil.ByteReverse(uid[:32])
	t.Index = jutil.GetUint32(uid[32:36])
}

func (t *TxInput) Deserialize(data []byte) {
	if len(data) != 36 {
		return
	}
	t.PrevHash = jutil.ByteReverse(data[:32])
	t.PrevIndex = jutil.GetUint32(data[32:36])
}

func GetTxInputUid(txHash []byte, index uint32) []byte {
	return GetTxHashIndexUid(txHash, index)
}

func GetTxInputsByHashes(txHashes [][]byte) ([]*TxInput, error) {
	var shardTxHashes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := uint32(GetShardByte(txHash))
		shardTxHashes[shard] = append(shardTxHashes[shard], jutil.ByteReverse(txHash))
	}
	var txInputs []*TxInput
	for shard, txHashes := range shardTxHashes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		err := db.GetByPrefixes(TopicTxInput, txHashes)
		if err != nil {
			return nil, jerr.Get("error getting db message tx inputs", err)
		}
		for _, msg := range db.Messages {
			var txInput = new(TxInput)
			txInput.SetUid(msg.Uid)
			txInput.Deserialize(msg.Message)
			txInputs = append(txInputs, txInput)
		}
	}
	return txInputs, nil
}

func GetTxInputs(outs []memo.Out) ([]*TxInput, error) {
	var shardOutGroups = make(map[uint32][]memo.Out)
	for _, out := range outs {
		shard := GetShardByte32(out.TxHash)
		shardOutGroups[shard] = append(shardOutGroups[shard], out)
	}
	var txInputs []*TxInput
	for shard, outGroup := range shardOutGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		var uids = make([][]byte, len(outGroup))
		for i := range outGroup {
			uids[i] = GetTxInputUid(outGroup[i].TxHash, outGroup[i].Index)
		}
		err := db.GetSpecific(TopicTxInput, uids)
		if err != nil {
			return nil, jerr.Get("error getting db", err)
		}
		for i := range db.Messages {
			var txInput = new(TxInput)
			txInput.SetUid(db.Messages[i].Uid)
			txInput.Deserialize(db.Messages[i].Message)
			txInputs = append(txInputs, txInput)
		}
	}
	return txInputs, nil
}

func GetTxInput(hash []byte, index uint32) (*TxInput, error) {
	shard := GetShardByte32(hash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	uid := GetTxInputUid(hash, index)
	if err := db.GetSingle(TopicTxInput, uid); err != nil {
		return nil, jerr.Get("error getting db", err)
	}
	if len(db.Messages) != 1 {
		return nil, nil
	}
	var txInput = new(TxInput)
	txInput.SetUid(db.Messages[0].Uid)
	txInput.Deserialize(db.Messages[0].Message)
	return txInput, nil
}
