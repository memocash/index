package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"github.com/memocash/server/ref/config"
	"sort"
	"strings"
	"time"
)

type HeightBlock struct {
	Height    int64
	BlockHash []byte
}

func (b HeightBlock) GetUid() []byte {
	return jutil.CombineBytes(jutil.GetInt64DataBig(b.Height), jutil.ByteReverse(b.BlockHash))
}

func (b HeightBlock) GetShard() uint {
	return uint(b.Height)
}

func (b HeightBlock) GetTopic() string {
	return TopicHeightBlock
}

func (b HeightBlock) Serialize() []byte {
	return nil
}

func (b *HeightBlock) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	b.Height = jutil.GetInt64Big(uid[:8])
	b.BlockHash = jutil.ByteReverse(uid[8:40])
}

func (b *HeightBlock) Deserialize([]byte) {}

func GetRecentHeightBlock(temp bool) (*HeightBlock, error) {
	var heightBlocks []*HeightBlock
	for i, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		err := dbClient.Get(TopicHeightBlock, client.GetMaxStart(), false)
		if err != nil {
			return nil, jerr.Getf(err, "error getting recent height block for shard: %d", i)
		}
		for i := range dbClient.Messages {
			var heightBlock = new(HeightBlock)
			heightBlock.SetUid(dbClient.Messages[i].Uid)
			heightBlock.Deserialize(dbClient.Messages[i].Message)
			heightBlocks = append(heightBlocks, heightBlock)
		}
	}
	if len(heightBlocks) == 0 {
		return nil, nil
	}
	var newestHeightBlock *HeightBlock
	for _, heightBlock := range heightBlocks {
		if newestHeightBlock == nil || newestHeightBlock.Height < heightBlock.Height {
			newestHeightBlock = heightBlock
		}
	}
	if newestHeightBlock == nil {
		return nil, nil
	}
	return newestHeightBlock, nil
}

func GetHeightBlockSingle(height int64) (*HeightBlock, error) {
	heightBlocks, err := GetHeightBlock(height)
	if err != nil {
		return nil, jerr.Get("error getting height block", err)
	}
	if len(heightBlocks) == 0 {
		return nil, jerr.Get("error no height blocks found", client.EntryNotFoundError)
	} else if len(heightBlocks) > 1 {
		var hashStrings = make([]string, len(heightBlocks))
		for i := range heightBlocks {
			hashStrings[i] = hs.GetTxString(heightBlocks[i].BlockHash)
		}
		return nil, jerr.Getf(client.MultipleEntryError, "error more than 1 height block found: %d (%s)",
			len(heightBlocks), strings.Join(hashStrings, ", "))
	}
	return heightBlocks[0], nil
}

func GetHeightBlock(height int64) ([]*HeightBlock, error) {
	shardConfig := config.GetShardConfig(uint32(height), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	err := dbClient.GetByPrefix(TopicHeightBlock, jutil.GetInt64DataBig(height))
	if err != nil {
		return nil, jerr.Get("error getting height blocks for height from queue client", err)
	}
	var heightBlocks = make([]*HeightBlock, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightBlocks[i] = new(HeightBlock)
		heightBlocks[i].SetUid(dbClient.Messages[i].Uid)
		heightBlocks[i].Deserialize(dbClient.Messages[i].Message)
	}
	return heightBlocks, nil
}

func GetHeightBlocks(shard uint32, startHeight int64, newest bool) ([]*HeightBlock, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startHeightBytes []byte
	if startHeight > 0 || !newest {
		startHeightBytes = jutil.GetInt64DataBig(startHeight)
	}
	err := dbClient.GetLarge(TopicHeightBlock, startHeightBytes, false, newest)
	if err != nil {
		return nil, jerr.Get("error getting height blocks from queue client", err)
	}
	var heightBlocks = make([]*HeightBlock, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightBlocks[i] = new(HeightBlock)
		heightBlocks[i].SetUid(dbClient.Messages[i].Uid)
		heightBlocks[i].Deserialize(dbClient.Messages[i].Message)
	}
	return heightBlocks, nil
}

func GetHeightBlocksAll(startHeight int64, waitSingle bool) ([]*HeightBlock, error) {
	var heightBlocks []*HeightBlock
	for _, shardConfig := range config.GetQueueShards() {
		if waitSingle && GetShard32(uint(startHeight+1)) != shardConfig.Min {
			continue
		}
		dbClient := client.NewClient(shardConfig.GetHost())
		var timeout time.Duration
		if waitSingle {
			timeout = time.Hour
		}
		err := dbClient.GetWOpts(client.Opts{
			Topic:   TopicHeightBlock,
			Start:   jutil.GetInt64DataBig(startHeight),
			Wait:    waitSingle,
			Max:     client.LargeLimit,
			Timeout: timeout,
		})
		if err != nil {
			return nil, jerr.Get("error getting height blocks from queue client all", err)
		}
		for i := range dbClient.Messages {
			var heightBlock = new(HeightBlock)
			heightBlock.SetUid(dbClient.Messages[i].Uid)
			heightBlock.Deserialize(dbClient.Messages[i].Message)
			heightBlocks = append(heightBlocks, heightBlock)
		}
	}
	sort.Slice(heightBlocks, func(i, j int) bool {
		return heightBlocks[i].Height < heightBlocks[j].Height
	})
	return heightBlocks, nil
}
