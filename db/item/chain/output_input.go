package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type OutputInput struct {
	PrevHash  [32]byte
	PrevIndex uint32
	Hash      [32]byte
	Index     uint32
}

func (t *OutputInput) GetTopic() string {
	return db.TopicChainOutputInput
}

func (t *OutputInput) GetShardSource() uint {
	return client.GenShardSource(t.PrevHash[:])
}

func (t *OutputInput) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.PrevHash[:]),
		jutil.GetUint32DataBig(t.PrevIndex),
		jutil.ByteReverse(t.Hash[:]),
		jutil.GetUint32DataBig(t.Index),
	)
}

func (t *OutputInput) SetUid(uid []byte) {
	if len(uid) != 72 {
		return
	}
	copy(t.PrevHash[:], jutil.ByteReverse(uid[:32]))
	t.PrevIndex = jutil.GetUint32Big(uid[32:36])
	copy(t.Hash[:], jutil.ByteReverse(uid[36:68]))
	t.Index = jutil.GetUint32Big(uid[68:72])
}

func (t *OutputInput) Serialize() []byte {
	return nil
}

func (t *OutputInput) Deserialize([]byte) {}

func GetOutputInputs(ctx context.Context, outs []memo.Out) ([]*OutputInput, error) {
	var shardPrefixes = make(map[uint32][]client.Prefix)
	for _, out := range outs {
		shard := db.GetShardIdFromByte32(out.TxHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], client.NewPrefix(jutil.CombineBytes(
			jutil.ByteReverse(out.TxHash),
			jutil.GetUint32DataBig(out.Index),
		)))
	}
	messages, err := db.GetByPrefixes(ctx, db.TopicChainOutputInput, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting by prefixes for chain output inputs; %w", err)
	}
	var outputInputs []*OutputInput
	for i := range messages {
		var outputInput = new(OutputInput)
		db.Set(outputInput, messages[i])
		outputInputs = append(outputInputs, outputInput)
	}
	return outputInputs, nil
}

func GetOutputInputsForTxHashes(ctx context.Context, txHashes [][32]byte) ([]*OutputInput, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicChainOutputInput, db.ShardPrefixesTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client messages memo output inputs; %w", err)
	}
	var outputInputs = make([]*OutputInput, len(messages))
	for i := range messages {
		outputInputs[i] = new(OutputInput)
		db.Set(outputInputs[i], messages[i])
	}
	return outputInputs, nil
}
