package chain

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"sort"
)

type TxBlock struct {
	TxHash    [32]byte
	BlockHash [32]byte
	Index     uint32
}

func (b *TxBlock) GetTopic() string {
	return db.TopicChainTxBlock
}

func (b *TxBlock) GetShard() uint {
	return client.GetByteShard(b.TxHash[:])
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

func GetTxBlockUid(txHash [32]byte, blockHash [32]byte) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash[:]), jutil.ByteReverse(blockHash[:]))
}

func GetTxBlocks(txHashes [][]byte) ([]*TxBlock, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardByte32(txHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(txHash))
	}
	wait := db.NewWait(len(shardPrefixes))
	var txBlocks []*TxBlock
	for shardT, prefixesT := range shardPrefixes {
		go func(shard uint32, prefixes [][]byte) {
			defer wait.Group.Done()
			sort.Slice(prefixes, func(i, j int) bool {
				return jutil.ByteLT(prefixes[i], prefixes[j])
			})
			dbClient := client.NewClient(config.GetShardConfig(shard, config.GetQueueShards()).GetHost())
			if err := dbClient.GetByPrefixes(db.TopicChainTxBlock, prefixes); err != nil {
				wait.AddError(jerr.Get("error getting client message chain tx blocks", err))
				return
			}
			wait.Lock.Lock()
			for _, msg := range dbClient.Messages {
				var txBlock = new(TxBlock)
				db.Set(txBlock, msg)
				txBlocks = append(txBlocks, txBlock)
			}
			wait.Lock.Unlock()
		}(shardT, prefixesT)
	}
	wait.Group.Wait()
	if len(wait.Errs) > 0 {
		return nil, jerr.Get("error getting chain tx blocks", jerr.Combine(wait.Errs...))
	}
	return txBlocks, nil
}
