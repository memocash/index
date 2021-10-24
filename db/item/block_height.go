package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
	"strings"
)

type BlockHeight struct {
	BlockHash []byte
	Height    int64
}

func (b BlockHeight) GetUid() []byte {
	return jutil.CombineBytes(jutil.ByteReverse(b.BlockHash), jutil.GetInt64DataBig(b.Height))
}

func (b BlockHeight) GetShard() uint {
	return client.GetByteShard(b.BlockHash)
}

func (b BlockHeight) GetTopic() string {
	return TopicBlockHeight
}

func (b BlockHeight) Serialize() []byte {
	return nil
}

func (b *BlockHeight) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	b.BlockHash = jutil.ByteReverse(uid[:32])
	b.Height = jutil.GetInt64Big(uid[32:40])
}

func (b *BlockHeight) Deserialize([]byte) {}

func GetBlockHeight(blockHash []byte) (*BlockHeight, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(blockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	err := dbClient.GetByPrefix(TopicBlockHeight, jutil.ByteReverse(blockHash))
	if err != nil {
		return nil, jerr.Get("error getting client message for block height", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no clients messages return for block height", client.EntryNotFoundError)
	} else if len(dbClient.Messages) > 1 {
		var hashStrings = make([]string, len(dbClient.Messages))
		for i := range dbClient.Messages {
			hashStrings[i] = dbClient.Messages[i].UidHex()
		}
		return nil, jerr.Newf("error more than 1 block height returned: %d (%s)",
			len(dbClient.Messages), strings.Join(hashStrings, ", "))
	}
	var blockHeight = new(BlockHeight)
	blockHeight.SetUid(dbClient.Messages[0].Uid)
	blockHeight.Deserialize(dbClient.Messages[0].Message)
	return blockHeight, nil
}
