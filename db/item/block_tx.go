package item

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
	"sort"
)

type BlockTx struct {
	TxHash    []byte
	BlockHash []byte
}

func (b BlockTx) GetUid() []byte {
	return GetBlockTxUid(b.BlockHash, b.TxHash)
}

func (b BlockTx) GetShard() uint {
	return client.GetByteShard(b.BlockHash)
}

func (b BlockTx) GetTopic() string {
	return TopicBlockTx
}

func (b BlockTx) Serialize() []byte {
	return nil
}

func (b *BlockTx) SetUid(uid []byte) {
	if len(uid) != 64 {
		return
	}
	b.BlockHash = jutil.ByteReverse(uid[:32])
	b.TxHash = jutil.ByteReverse(uid[32:64])
}

func (b *BlockTx) Deserialize([]byte) {}

func GetBlockTxUid(blockHash, txHash []byte) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(blockHash), jutil.ByteReverse(txHash))
}

func GetBlockTx(blockHash, txHash []byte) (*BlockTx, error) {
	shard := client.GetByteShard32(blockHash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	err := dbClient.GetSingle(TopicBlockTx, GetBlockTxUid(blockHash, txHash))
	if err != nil {
		return nil, jerr.Get("error getting client message block tx single", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of block tx client messages: %d", len(dbClient.Messages))
	}
	var block = new(BlockTx)
	block.SetUid(dbClient.Messages[0].Uid)
	return block, nil
}

type BlockTxesRequest struct {
	BlockHash []byte
	StartUid  []byte
	Limit     uint32
}

func GetBlockTxes(request BlockTxesRequest) ([]*BlockTx, error) {
	shard := client.GetByteShard32(request.BlockHash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var limit uint32
	if request.Limit > 0 {
		limit = request.Limit
	} else {
		limit = client.LargeLimit
	}
	err := dbClient.GetWOpts(client.Opts{
		Topic:    TopicBlockTx,
		Prefixes: [][]byte{jutil.ByteReverse(request.BlockHash)},
		Start:    request.StartUid,
		Max:      limit,
	})
	if err != nil {
		return nil, jerr.Get("error getting client message", err)
	}
	var blocks = make([]*BlockTx, len(dbClient.Messages))
	for i := range dbClient.Messages {
		blocks[i] = new(BlockTx)
		blocks[i].SetUid(dbClient.Messages[i].Uid)
	}
	sort.Slice(blocks, func(i, j int) bool {
		return bytes.Compare(blocks[i].BlockHash, blocks[j].BlockHash) == -1
	})
	return blocks, nil
}

func GetBlockTxCount(blockHash []byte) (uint64, error) {
	shard := client.GetByteShard32(blockHash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	count, err := dbClient.GetTopicCount(TopicBlockTx, blockHash)
	if err != nil {
		return 0, jerr.Get("error getting block tx count", err)
	}
	return count, nil
}
