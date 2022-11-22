package chain

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"sort"
)

type BlockTx struct {
	BlockHash [32]byte
	TxHash    [32]byte
	Index     uint32
}

func (b *BlockTx) GetTopic() string {
	return db.TopicChainBlockTx
}

func (b *BlockTx) GetShard() uint {
	return client.GetByteShard(b.BlockHash[:])
}

func (b *BlockTx) GetUid() []byte {
	return GetBlockTxUid(b.BlockHash[:], b.TxHash[:])
}

func (b *BlockTx) SetUid(uid []byte) {
	if len(uid) != 64 {
		return
	}
	copy(b.BlockHash[:], jutil.ByteReverse(uid[:32]))
	copy(b.TxHash[:], jutil.ByteReverse(uid[32:64]))
}

func (b *BlockTx) Serialize() []byte {
	return jutil.GetUint32DataBig(b.Index)
}

func (b *BlockTx) Deserialize(data []byte) {
	if len(data) != 4 {
		return
	}
	b.Index = jutil.GetUint32Big(data[:4])
}

func GetBlockTxUid(blockHash, txHash []byte) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(blockHash), jutil.ByteReverse(txHash))
}

func GetBlockTx(blockHash, txHash []byte) (*BlockTx, error) {
	shard := client.GetByteShard32(blockHash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicChainBlockTx, GetBlockTxUid(blockHash, txHash)); err != nil {
		return nil, jerr.Get("error getting client message chain block tx single", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of chain block tx client messages: %d", len(dbClient.Messages))
	}
	var block = new(BlockTx)
	db.Set(block, dbClient.Messages[0])
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
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicChainBlockTx,
		Prefixes: [][]byte{jutil.ByteReverse(request.BlockHash)},
		Start:    request.StartUid,
		Max:      limit,
	}); err != nil {
		return nil, jerr.Get("error getting client message", err)
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
