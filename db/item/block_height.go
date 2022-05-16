package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"sort"
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

func GetBlockHeights(blockHashes [][]byte) ([]*BlockHeight, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, blockHash := range blockHashes {
		shard := GetShardByte32(blockHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(blockHash))
	}
	wait := NewWait(len(shardPrefixes))
	var blockHeights []*BlockHeight
	for shardT, prefixesT := range shardPrefixes {
		go func(shard uint32, prefixes [][]byte) {
			defer wait.Group.Done()
			sort.Slice(prefixes, func(i, j int) bool {
				return jutil.ByteLT(prefixes[i], prefixes[j])
			})
			db := client.NewClient(config.GetShardConfig(shard, config.GetQueueShards()).GetHost())
			if err := db.GetByPrefixes(TopicBlockHeight, prefixes); err != nil {
				wait.AddError(jerr.Get("error getting client message block heights", err))
				return
			}
			wait.Lock.Lock()
			for _, msg := range db.Messages {
				var blockHeight = new(BlockHeight)
				blockHeight.SetUid(msg.Uid)
				blockHeight.Deserialize(msg.Message)
				blockHeights = append(blockHeights, blockHeight)
			}
			wait.Lock.Unlock()
		}(shardT, prefixesT)
	}
	wait.Group.Wait()
	if len(wait.Errs) > 0 {
		return nil, jerr.Get("error getting block heights", jerr.Combine(wait.Errs...))
	}
	return blockHeights, nil
}

func ListenBlockHeights() (chan *BlockHeight, error) {
	var chanBlockHeight = make(chan *BlockHeight)
	for _, shardConfig := range config.GetQueueShards() {
		db := client.NewClient(shardConfig.GetHost())
		chanMessage, err := db.Listen(TopicBlockHeight, nil)
		if err != nil {
			return nil, jerr.Get("error getting block height listen message chan", err)
		}
		go func() {
			for {
				msg := <-chanMessage
				if msg == nil {
					chanBlockHeight <- nil
					close(chanBlockHeight)
					return
				}
				var blockHeight = new(BlockHeight)
				Set(blockHeight, *msg)
				chanBlockHeight <- blockHeight
			}
		}()
	}
	return chanBlockHeight, nil
}
