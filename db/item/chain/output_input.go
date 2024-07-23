package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
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

func GetOutputInput(out memo.Out) ([]*OutputInput, error) {
	shard := db.GetShardIdFromByte32(out.TxHash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	prefix := jutil.CombineBytes(jutil.ByteReverse(out.TxHash), jutil.GetUint32Data(out.Index))
	if err := dbClient.GetByPrefix(db.TopicChainOutputInput, prefix); err != nil {
		return nil, fmt.Errorf("error getting by prefix for chain output input; %w", err)
	}
	var outputInputs = make([]*OutputInput, len(dbClient.Messages))
	for i := range dbClient.Messages {
		outputInputs[i] = new(OutputInput)
		db.Set(outputInputs[i], dbClient.Messages[i])
	}
	return outputInputs, nil
}

func GetOutputInputs(ctx context.Context, outs []memo.Out) ([]*OutputInput, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, out := range outs {
		shard := db.GetShardIdFromByte32(out.TxHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.CombineBytes(
			jutil.ByteReverse(out.TxHash),
			jutil.GetUint32DataBig(out.Index),
		))
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

func GetOutputInputsForTxHashes(ctx context.Context, txHashes [][]byte) ([]*OutputInput, error) {
	var shardOutGroups = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardIdFromByte32(txHash)
		shardOutGroups[shard] = append(shardOutGroups[shard], txHash)
	}
	var outputInputs []*OutputInput
	for shard, outGroup := range shardOutGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		var prefixes = make([]client.Prefix, len(outGroup))
		for i := range outGroup {
			prefixes[i] = client.Prefix{Prefix: jutil.ByteReverse(outGroup[i])}
		}
		if err := dbClient.GetByPrefixesNew(ctx, db.TopicChainOutputInput, prefixes); err != nil {
			return nil, fmt.Errorf("error getting by prefixes for chain output inputs by tx hashes; %w", err)
		}
		for i := range dbClient.Messages {
			var outputInput = new(OutputInput)
			db.Set(outputInput, dbClient.Messages[i])
			outputInputs = append(outputInputs, outputInput)
		}
	}
	return outputInputs, nil
}
