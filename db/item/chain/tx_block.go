package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type TxBlock struct {
	TxHash    [32]byte
	BlockHash [32]byte
	Index     uint32
}

func (b *TxBlock) GetTopic() string {
	return db.TopicChainTxBlock
}

func (b *TxBlock) GetShardSource() uint {
	return client.GenShardSource(b.TxHash[:])
}

func (b *TxBlock) GetUid() []byte {
	return GetTxBlockUid(b.TxHash, b.BlockHash)
}

func (b *TxBlock) SetUid(uid []byte) {
	if len(uid) != 64 {
		return
	}
	copy(b.TxHash[:], jutil.ByteReverse(uid[:32]))
	copy(b.BlockHash[:], jutil.ByteReverse(uid[32:64]))
}

func (b *TxBlock) Serialize() []byte {
	return nil
}

func (b *TxBlock) Deserialize([]byte) {}

func GetTxBlockUid(txHash, blockHash [32]byte) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash[:]), jutil.ByteReverse(blockHash[:]))
}

func GetSingleTxBlock(txHash, blockHash [32]byte) (*TxBlock, error) {
	shardConfig := config.GetShardConfig(client.GenShardSource32(txHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicChainTxBlock, GetTxBlockUid(txHash, blockHash)); err != nil {
		return nil, fmt.Errorf("error getting client message single tx block; %w", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, fmt.Errorf("error unexpected number of single tx block client messages: %d", len(dbClient.Messages))
	}
	var txBlock = new(TxBlock)
	db.Set(txBlock, dbClient.Messages[0])
	return txBlock, nil
}

func GetSingleTxBlocks(ctx context.Context, txHash [32]byte) ([]*TxBlock, error) {
	shardConfig := config.GetShardConfig(client.GenShardSource32(txHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	prefix := client.NewPrefix(jutil.ByteReverse(txHash[:]))
	if err := dbClient.GetByPrefix(ctx, db.TopicChainTxBlock, prefix); err != nil {
		return nil, fmt.Errorf("error getting client message chain tx block by prefix; %w", err)
	}
	var txBlocks []*TxBlock
	for _, msg := range dbClient.Messages {
		var txBlock = new(TxBlock)
		db.Set(txBlock, msg)
		txBlocks = append(txBlocks, txBlock)
	}
	return txBlocks, nil
}

func GetTxBlocks(ctx context.Context, txHashes [][32]byte) ([]*TxBlock, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicChainTxBlock, db.ShardPrefixesTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client message chain tx blocks; %w", err)
	}
	var txBlocks []*TxBlock
	for _, msg := range messages {
		var txBlock = new(TxBlock)
		db.Set(txBlock, msg)
		txBlocks = append(txBlocks, txBlock)
	}
	return txBlocks, nil
}
