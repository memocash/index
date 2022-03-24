package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"strings"
)

type BlockHeightRaw struct {
	BlockHash []byte
	Height    int64
}

func (b BlockHeightRaw) GetUid() []byte {
	return jutil.CombineBytes(jutil.ByteReverse(b.BlockHash), jutil.GetInt64DataBig(b.Height))
}

func (b BlockHeightRaw) GetShard() uint {
	return client.GetByteShard(b.BlockHash)
}

func (b BlockHeightRaw) GetTopic() string {
	return TopicBlockHeightRaw
}

func (b BlockHeightRaw) Serialize() []byte {
	return nil
}

func (b *BlockHeightRaw) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	b.BlockHash = jutil.ByteReverse(uid[:32])
	b.Height = jutil.GetInt64Big(uid[32:40])
}

func (b *BlockHeightRaw) Deserialize([]byte) {}

func GetBlockHeightRaw(blockHash []byte) (*BlockHeightRaw, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(blockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetByPrefix(TopicBlockHeightRaw, jutil.ByteReverse(blockHash)); err != nil {
		return nil, jerr.Get("error getting client message for block height raw", err)
	}
	if len(db.Messages) == 0 {
		return nil, jerr.Get("error no clients messages return for block height raw", client.EntryNotFoundError)
	} else if len(db.Messages) > 1 {
		var hashStrings = make([]string, len(db.Messages))
		for i := range db.Messages {
			hashStrings[i] = db.Messages[i].UidHex()
		}
		return nil, jerr.Newf("error more than 1 block height raw returned: %d (%s)",
			len(db.Messages), strings.Join(hashStrings, ", "))
	}
	var blockHeightRaw = new(BlockHeightRaw)
	Set(blockHeightRaw, db.Messages[0])
	return blockHeightRaw, nil
}

func GetBlockHeightRaws(blockHashes [][]byte) ([]*BlockHeightRaw, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, blockHash := range blockHashes {
		shard := GetShardByte32(blockHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(blockHash))
	}
	var blockHeightRaws []*BlockHeightRaw
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetByPrefixes(TopicBlockHeightRaw, prefixes); err != nil {
			return nil, jerr.Get("error getting client message block height raws", err)
		}
		for i := range db.Messages {
			var blockHeightRaw = new(BlockHeightRaw)
			Set(blockHeightRaw, db.Messages[i])
			blockHeightRaws = append(blockHeightRaws, blockHeightRaw)
		}
	}
	return blockHeightRaws, nil
}
