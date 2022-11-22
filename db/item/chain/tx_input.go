package chain

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type TxInput struct {
	TxHash       [32]byte
	Index        uint32
	PrevHash     [32]byte
	PrevIndex    uint32
	Sequence     uint32
	UnlockScript []byte
}

func (t *TxInput) GetTopic() string {
	return db.TopicChainTxInput
}

func (t *TxInput) GetShard() uint {
	return client.GetByteShard(t.TxHash[:])
}

func (t *TxInput) GetUid() []byte {
	return GetTxInputUid(t.TxHash, t.Index)
}

func (t *TxInput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	copy(t.TxHash[:], jutil.ByteReverse(uid[:32]))
	t.Index = jutil.GetUint32(uid[32:36])
}

func (t *TxInput) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.PrevHash[:]),
		jutil.GetUint32DataBig(t.PrevIndex),
		jutil.GetUint32Data(t.Sequence),
		t.UnlockScript,
	)
}

func (t *TxInput) Deserialize(data []byte) {
	if len(data) < 40 {
		return
	}
	copy(t.PrevHash[:], jutil.ByteReverse(data[:32]))
	t.PrevIndex = jutil.GetUint32Big(data[32:36])
	t.Sequence = jutil.GetUint32(data[36:40])
	t.UnlockScript = data[40:]
}

func GetTxInputUid(txHash [32]byte, index uint32) []byte {
	return db.GetTxHashIndexUid(txHash[:], index)
}

func GetTxInputsByHashes(txHashes [][]byte) ([]*TxInput, error) {
	var shardTxHashes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := uint32(db.GetShardByte(txHash))
		shardTxHashes[shard] = append(shardTxHashes[shard], jutil.ByteReverse(txHash))
	}
	var txInputs []*TxInput
	for shard, txHashes := range shardTxHashes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		err := dbClient.GetByPrefixes(db.TopicChainTxInput, txHashes)
		if err != nil {
			return nil, jerr.Get("error getting db message chain tx inputs", err)
		}
		for _, msg := range dbClient.Messages {
			var txInput = new(TxInput)
			db.Set(txInput, msg)
			txInputs = append(txInputs, txInput)
		}
	}
	return txInputs, nil
}
