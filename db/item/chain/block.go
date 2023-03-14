package chain

import (
	"github.com/jchavannes/jgo/jerr"
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

func (b *Block) GetShard() uint {
	return client.GetByteShard(b.Hash[:])
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
	shardConfig := config.GetShardConfig(client.GetByteShard32(blockHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicChainBlock, jutil.ByteReverse(blockHash[:])); err != nil {
		return nil, jerr.Get("error getting client message block", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of messages: %d", len(dbClient.Messages))
	}
	var block = new(Block)
	db.Set(block, dbClient.Messages[0])
	return block, nil
}

func GetBlocks(blockHashes [][32]byte) ([]*Block, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, blockHash := range blockHashes {
		shard := db.GetShardByte32(blockHash[:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(blockHash[:]))
	}
	messages, err := db.GetSpecific(db.TopicChainBlock, shardUids)
	if err != nil {
		return nil, jerr.Get("error getting client message chain blocks", err)
	}
	var blocks []*Block
	for _, msg := range messages {
		var block = new(Block)
		db.Set(block, msg)
		blocks = append(blocks, block)
	}
	return blocks, nil
}
