package chain

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type BlockInfo struct {
	BlockHash [32]byte
	Size      int64
	TxCount   int
}

func (b *BlockInfo) GetTopic() string {
	return db.TopicChainBlockInfo
}

func (b *BlockInfo) GetShard() uint {
	return client.GetByteShard(b.BlockHash[:])
}

func (b *BlockInfo) GetUid() []byte {
	return jutil.ByteReverse(b.BlockHash[:])
}

func (b *BlockInfo) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	copy(b.BlockHash[:], jutil.ByteReverse(uid[:32]))
}

func (b *BlockInfo) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.GetInt64Data(b.Size),
		jutil.GetIntData(b.TxCount),
	)
}

func (b *BlockInfo) Deserialize(data []byte) {
	if len(data) != 12 {
		return
	}
	b.Size = jutil.GetInt64(data[:8])
	b.TxCount = jutil.GetInt(data[8:12])
}

func GetBlockInfo(blockHash []byte) (*BlockInfo, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(blockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicChainBlockInfo, jutil.ByteReverse(blockHash)); err != nil {
		return nil, jerr.Get("error getting client message block info", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of messages for block info: %d", len(dbClient.Messages))
	}
	var blockInfo = new(BlockInfo)
	db.Set(blockInfo, dbClient.Messages[0])
	return blockInfo, nil
}

func GetBlockInfos(blockHashes [][]byte) ([]*BlockInfo, error) {
	var shardBlockHashGroups = make(map[uint32][][]byte)
	for _, blockHash := range blockHashes {
		shard := db.GetShardByte32(blockHash)
		shardBlockHashGroups[shard] = append(shardBlockHashGroups[shard], jutil.ByteReverse(blockHash))
	}
	var blocks []*BlockInfo
	for shard, blockHashGroup := range shardBlockHashGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetSpecific(db.TopicChainBlockInfo, blockHashGroup); err != nil {
			return nil, jerr.Get("error getting client message chain block infos", err)
		}
		for _, msg := range dbClient.Messages {
			var block = new(BlockInfo)
			db.Set(block, msg)
			blocks = append(blocks, block)
		}
	}
	return blocks, nil
}
