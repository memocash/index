package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type Block struct {
	Hash []byte
	Raw  []byte
}

func (b Block) GetUid() []byte {
	return jutil.ByteReverse(b.Hash)
}

func (b Block) GetShard() uint {
	return client.GetByteShard(b.Hash)
}

func (b Block) GetTopic() string {
	return db.TopicBlock
}

func (b Block) Serialize() []byte {
	return b.Raw
}

func (b *Block) SetUid(uid []byte) {
	b.Hash = jutil.ByteReverse(uid)
}

func (b *Block) Deserialize(data []byte) {
	b.Raw = data

}

func GetBlock(blockHash []byte) (*Block, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(blockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicBlock, jutil.ByteReverse(blockHash)); err != nil {
		return nil, jerr.Get("error getting client message block", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of messages: %d", len(dbClient.Messages))
	}
	var block = new(Block)
	db.Set(block, dbClient.Messages[0])
	return block, nil
}

func GetBlocks(blockHashes [][]byte) ([]*Block, error) {
	var shardBlockHashGroups = make(map[uint32][][]byte)
	for _, blockHash := range blockHashes {
		shard := db.GetShardByte32(blockHash)
		shardBlockHashGroups[shard] = append(shardBlockHashGroups[shard], jutil.ByteReverse(blockHash))
	}
	var blocks []*Block
	for shard, blockHashGroup := range shardBlockHashGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetSpecific(db.TopicBlock, blockHashGroup); err != nil {
			return nil, jerr.Get("error getting client message blocks", err)
		}
		for _, msg := range dbClient.Messages {
			var block = new(Block)
			db.Set(block, msg)
			blocks = append(blocks, block)
		}
	}
	return blocks, nil
}
