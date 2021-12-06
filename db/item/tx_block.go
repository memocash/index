package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

type TxBlock struct {
	TxHash    []byte
	BlockHash []byte
}

func (b TxBlock) GetUid() []byte {
	return GetTxBlockUid(b.TxHash, b.BlockHash)
}

func (b TxBlock) GetShard() uint {
	return client.GetByteShard(b.TxHash)
}

func (b TxBlock) GetTopic() string {
	return TopicTxBlock
}

func (b TxBlock) Serialize() []byte {
	return nil
}

func (b *TxBlock) SetUid(uid []byte) {
	if len(uid) != 64 {
		return
	}
	b.TxHash = jutil.ByteReverse(uid[:32])
	b.BlockHash = jutil.ByteReverse(uid[32:64])
}

func (b *TxBlock) Deserialize([]byte) {}

func GetTxBlockUid(txHash []byte, blockHash []byte) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash), jutil.ByteReverse(blockHash))
}

func GetSingleTxBlock(txHash, blockHash []byte) (*TxBlock, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(txHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	err := db.GetSingle(TopicTxBlock, GetTxBlockUid(txHash, blockHash))
	if err != nil {
		return nil, jerr.Get("error getting client message single tx block", err)
	}
	if len(db.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of single tx block client messages: %d", len(db.Messages))
	}
	var txBlock = new(TxBlock)
	txBlock.SetUid(db.Messages[0].Uid)
	txBlock.Deserialize(db.Messages[0].Message)
	return txBlock, nil
}

func GetSingleTxBlocks(txHash []byte) ([]*TxBlock, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(txHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	err := db.GetByPrefix(TopicTxBlock, jutil.ByteReverse(txHash))
	if err != nil {
		return nil, jerr.Get("error getting client message tx block by prefix", err)
	}
	var txBlocks []*TxBlock
	for _, msg := range db.Messages {
		var txBlock = new(TxBlock)
		txBlock.SetUid(msg.Uid)
		txBlocks = append(txBlocks, txBlock)
	}
	return txBlocks, nil
}

func GetTxBlocks(txHashes [][]byte) ([]*TxBlock, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := GetShardByte32(txHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(txHash))
	}
	var txBlocks []*TxBlock
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		err := db.GetByPrefixes(TopicTxBlock, prefixes)
		if err != nil {
			return nil, jerr.Get("error getting client message tx blocks", err)
		}
		for _, msg := range db.Messages {
			var txBlock = new(TxBlock)
			txBlock.SetUid(msg.Uid)
			txBlock.Deserialize(msg.Message)
			txBlocks = append(txBlocks, txBlock)
		}
	}
	return txBlocks, nil
}
