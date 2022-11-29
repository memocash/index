package chain

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type Tx struct {
	TxHash   [32]byte
	Version  int32
	LockTime uint32
}

func (t *Tx) GetTopic() string {
	return db.TopicChainTx
}

func (t *Tx) GetShard() uint {
	return client.GetByteShard(t.TxHash[:])
}

func (t *Tx) GetUid() []byte {
	return jutil.ByteReverse(t.TxHash[:])
}

func (t *Tx) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	copy(t.TxHash[:], jutil.ByteReverse(uid))
}

func (t *Tx) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.GetInt32Data(t.Version),
		jutil.GetUint32Data(t.LockTime),
	)
}

func (t *Tx) Deserialize(data []byte) {
	if len(data) < 8 {
		return
	}
	t.Version = jutil.GetInt32(data[:4])
	t.LockTime = jutil.GetUint32(data[4:8])
}

func GetTxsByHashes(txHashes [][32]byte) ([]*Tx, error) {
	var shardTxHashes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := uint32(db.GetShardByte(txHash[:]))
		shardTxHashes[shard] = append(shardTxHashes[shard], jutil.ByteReverse(txHash[:]))
	}
	var txs []*Tx
	for shard, txHashes := range shardTxHashes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		err := dbClient.GetByPrefixes(db.TopicChainTx, txHashes)
		if err != nil {
			return nil, jerr.Get("error getting db message chain txs by hashes", err)
		}
		for _, msg := range dbClient.Messages {
			var tx = new(Tx)
			db.Set(tx, msg)
			txs = append(txs, tx)
		}
	}
	return txs, nil
}
