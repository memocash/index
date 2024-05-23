package chain

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"sort"
)

type BlockTx struct {
	BlockHash [32]byte
	Index     uint32
	TxHash    [32]byte
}

func (b *BlockTx) GetTopic() string {
	return db.TopicChainBlockTx
}

func (b *BlockTx) GetShardSource() uint {
	return client.GenShardSource(b.BlockHash[:])
}

func (b *BlockTx) GetUid() []byte {
	return GetBlockTxUid(b.BlockHash, b.Index)
}

func (b *BlockTx) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	copy(b.BlockHash[:], jutil.ByteReverse(uid[:32]))
	b.Index = jutil.GetUint32Big(uid[32:36])
}

func (b *BlockTx) Serialize() []byte {
	return jutil.ByteReverse(b.TxHash[:])
}

func (b *BlockTx) Deserialize(data []byte) {
	if len(data) != 32 {
		return
	}
	copy(b.TxHash[:], jutil.ByteReverse(data[:32]))
}

func GetBlockTxUid(blockHash [32]byte, index uint32) []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(blockHash[:]),
		jutil.GetUint32DataBig(index),
	)
}

func GetBlockTx(blockHash [32]byte, index uint32) (*BlockTx, error) {
	shard := client.GenShardSource32(blockHash[:])
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicChainBlockTx, GetBlockTxUid(blockHash, index)); err != nil {
		return nil, fmt.Errorf("error getting client message chain block tx single; %w", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, fmt.Errorf("error unexpected number of chain block tx client messages: %d", len(dbClient.Messages))
	}
	var block = new(BlockTx)
	db.Set(block, dbClient.Messages[0])
	return block, nil
}

type BlockTxsRequest struct {
	Context    context.Context
	BlockHash  [32]byte
	StartIndex uint32
	Limit      uint32
}

func GetBlockTxs(request BlockTxsRequest) ([]*BlockTx, error) {
	shard := client.GenShardSource32(request.BlockHash[:])
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var limit uint32
	if request.Limit > 0 {
		limit = request.Limit
	} else {
		limit = client.LargeLimit
	}
	var startUid []byte
	if request.StartIndex > 0 {
		startUid = GetBlockTxUid(request.BlockHash, request.StartIndex)
	}
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicChainBlockTx,
		Prefixes: [][]byte{jutil.ByteReverse(request.BlockHash[:])},
		Start:    startUid,
		Max:      limit,
		Context:  request.Context,
	}); err != nil {
		return nil, fmt.Errorf("error getting client message; %w", err)
	}
	var blocks = make([]*BlockTx, len(dbClient.Messages))
	for i := range dbClient.Messages {
		blocks[i] = new(BlockTx)
		db.Set(blocks[i], dbClient.Messages[i])
	}
	sort.Slice(blocks, func(i, j int) bool {
		return bytes.Compare(blocks[i].BlockHash[:], blocks[j].BlockHash[:]) == -1
	})
	return blocks, nil
}
