package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type Block struct {
	Hash [32]byte
	Raw  []byte
}

func (b *Block) GetTopic() string {
	return db.TopicChainBlock
}

func (b *Block) GetShardSource() uint {
	return client.GenShardSource(b.Hash[:])
}

func (b *Block) GetUid() []byte {
	return jutil.ByteReverse(b.Hash[:])
}

func (b *Block) SetUid(uid []byte) {
	copy(b.Hash[:], jutil.ByteReverse(uid))
}

func (b *Block) Serialize() []byte {
	return b.Raw
}

func (b *Block) Deserialize(data []byte) {
	b.Raw = data
}

func GetBlock(blockHash [32]byte) (*Block, error) {
	shardConfig := config.GetShardConfig(client.GenShardSource32(blockHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicChainBlock, jutil.ByteReverse(blockHash[:])); err != nil {
		return nil, fmt.Errorf("error getting client message block; %w", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, fmt.Errorf("error unexpected number of messages: %d", len(dbClient.Messages))
	}
	var block = new(Block)
	db.Set(block, dbClient.Messages[0])
	return block, nil
}

func GetBlocks(ctx context.Context, blockHashes [][32]byte) ([]*Block, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, blockHash := range blockHashes {
		shard := db.GetShardIdFromByte32(blockHash[:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(blockHash[:]))
	}
	messages, err := db.GetSpecific(ctx, db.TopicChainBlock, shardUids)
	if err != nil {
		return nil, fmt.Errorf("error getting client message chain blocks; %w", err)
	}
	var blocks []*Block
	for _, msg := range messages {
		var block = new(Block)
		db.Set(block, msg)
		blocks = append(blocks, block)
	}
	return blocks, nil
}
