package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type Tx struct {
	TxHash   [32]byte
	Version  int32
	LockTime uint32
}

func (t *Tx) GetTopic() string {
	return db.TopicChainTx
}

func (t *Tx) GetShardSource() uint {
	return client.GenShardSource(t.TxHash[:])
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

func GetTxsByHashes(ctx context.Context, txHashes [][32]byte) ([]*Tx, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicChainTx, db.ShardPrefixesTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting db message chain txs by hashes; %w", err)
	}
	var txs = make([]*Tx, len(messages))
	for i := range messages {
		txs[i] = new(Tx)
		db.Set(txs[i], messages[i])
	}
	return txs, nil
}
