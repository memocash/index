package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type OutputInputSingle struct {
	PrevHash  [32]byte
	PrevIndex uint32
	Hash      [32]byte
	Index     uint32
}

func (t *OutputInputSingle) GetTopic() string {
	return db.TopicChainOutputInputSingle
}

func (t *OutputInputSingle) GetShardSource() uint {
	return client.GenShardSource(t.PrevHash[:])
}

func (t *OutputInputSingle) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.PrevHash[:]),
		jutil.GetUint32DataBig(t.PrevIndex),
	)
}

func (t *OutputInputSingle) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	copy(t.PrevHash[:], jutil.ByteReverse(uid[:32]))
	t.PrevIndex = jutil.GetUint32Big(uid[32:36])
}

func (t *OutputInputSingle) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.Hash[:]),
		jutil.GetUint32DataBig(t.Index),
	)
}

func (t *OutputInputSingle) Deserialize(data []byte) {
	if len(data) != 36 {
		return
	}
	copy(t.Hash[:], jutil.ByteReverse(data[:32]))
	t.Index = jutil.GetUint32Big(data[32:36])
}

func GetOutputInputSingles(ctx context.Context, outs []memo.Out) ([]*OutputInputSingle, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, out := range outs {
		shard := db.GetShardIdFromByte32(out.TxHash)
		shardUids[shard] = append(shardUids[shard], jutil.CombineBytes(
			jutil.ByteReverse(out.TxHash),
			jutil.GetUint32DataBig(out.Index),
		))
	}
	messages, err := db.GetSpecific(ctx, db.TopicChainOutputInputSingle, shardUids)
	if err != nil {
		return nil, fmt.Errorf("error getting specific for chain output input singles; %w", err)
	}
	var outputInputSingles []*OutputInputSingle
	for i := range messages {
		var outputInputSingle = new(OutputInputSingle)
		db.Set(outputInputSingle, messages[i])
		outputInputSingles = append(outputInputSingles, outputInputSingle)
	}
	return outputInputSingles, nil
}

func GetOutputInputSinglesForTxHashes(ctx context.Context, txHashes [][32]byte) ([]*OutputInputSingle, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicChainOutputInputSingle, db.ShardPrefixesTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client messages output input singles; %w", err)
	}
	var outputInputSingles = make([]*OutputInputSingle, len(messages))
	for i := range messages {
		outputInputSingles[i] = new(OutputInputSingle)
		db.Set(outputInputSingles[i], messages[i])
	}
	return outputInputSingles, nil
}
