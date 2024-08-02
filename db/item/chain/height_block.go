package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"sort"
	"strings"
)

type HeightBlock struct {
	Height    int64
	BlockHash [32]byte
}

func (b *HeightBlock) GetTopic() string {
	return db.TopicChainHeightBlock
}

func (b *HeightBlock) GetShardSource() uint {
	return uint(b.Height)
}

func (b *HeightBlock) GetUid() []byte {
	return jutil.CombineBytes(jutil.GetInt64DataBig(b.Height), jutil.ByteReverse(b.BlockHash[:]))
}

func (b *HeightBlock) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	b.Height = jutil.GetInt64Big(uid[:8])
	copy(b.BlockHash[:], jutil.ByteReverse(uid[8:40]))
}

func (b *HeightBlock) Serialize() []byte {
	return nil
}

func (b *HeightBlock) Deserialize([]byte) {}

func GetRecentHeightBlock(ctx context.Context) (*HeightBlock, error) {
	var heightBlocks []*HeightBlock
	for i, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetLast(ctx, db.TopicChainHeightBlock); err != nil {
			return nil, fmt.Errorf("error getting recent height block for shard: %d; %w", i, err)
		}
		for i := range dbClient.Messages {
			var heightBlock = new(HeightBlock)
			db.Set(heightBlock, dbClient.Messages[i])
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

func GetHeightBlockSingle(ctx context.Context, height int64) (*HeightBlock, error) {
	heightBlocks, err := GetHeightBlock(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("error getting height block; %w", err)
	}
	if len(heightBlocks) == 0 {
		return nil, fmt.Errorf("error no height blocks found; %w", client.EntryNotFoundError)
	} else if len(heightBlocks) > 1 {
		var hashStrings = make([]string, len(heightBlocks))
		for i := range heightBlocks {
			hashStrings[i] = chainhash.Hash(heightBlocks[i].BlockHash).String()
		}
		return nil, fmt.Errorf("error more than 1 height block found: %d (%s); %w",
			len(heightBlocks), strings.Join(hashStrings, ", "), client.MultipleEntryError)
	}
	return heightBlocks[0], nil
}

func GetHeightBlock(ctx context.Context, height int64) ([]*HeightBlock, error) {
	shardConfig := config.GetShardConfig(uint32(height), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	prefix := client.NewPrefix(jutil.GetInt64DataBig(height))
	if err := dbClient.GetByPrefix(ctx, db.TopicChainHeightBlock, prefix); err != nil {
		return nil, fmt.Errorf("error getting height blocks for height from queue client; %w", err)
	}
	var heightBlocks = make([]*HeightBlock, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightBlocks[i] = new(HeightBlock)
		db.Set(heightBlocks[i], dbClient.Messages[i])
	}
	return heightBlocks, nil
}

func GetHeightBlocks(ctx context.Context, shard uint32, startHeight int64, desc bool) ([]*HeightBlock, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startHeightBytes []byte
	if startHeight > 0 || !desc {
		startHeightBytes = jutil.GetInt64DataBig(startHeight)
	}
	if err := dbClient.GetByPrefix(ctx, db.TopicChainHeightBlock, client.Prefix{
		Start: startHeightBytes,
		Limit: client.LargeLimit,
	}, client.NewOptionOrder(desc)); err != nil {
		return nil, fmt.Errorf("error getting height blocks from queue client; %w", err)
	}
	var heightBlocks = make([]*HeightBlock, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightBlocks[i] = new(HeightBlock)
		db.Set(heightBlocks[i], dbClient.Messages[i])
	}
	return heightBlocks, nil
}

func GetHeightBlocksAll(ctx context.Context, startHeight int64) ([]*HeightBlock, error) {
	heightBlocks, err := GetHeightBlocksAllLimit(ctx, startHeight, client.LargeLimit, false)
	if err != nil {
		return nil, fmt.Errorf("error getting height blocks all large limit; %w", err)
	}
	return heightBlocks, nil
}

func GetHeightBlocksAllDefault(ctx context.Context, startHeight int64, newest bool) ([]*HeightBlock, error) {
	heightBlocks, err := GetHeightBlocksAllLimit(ctx, startHeight, client.DefaultLimit, newest)
	if err != nil {
		return nil, fmt.Errorf("error getting height blocks all default limit; %w", err)
	}
	return heightBlocks, nil
}

func GetHeightBlocksAllLimit(ctx context.Context, startHeight int64, limit int, newest bool) ([]*HeightBlock, error) {
	var heightBlocks []*HeightBlock
	shardConfigs := config.GetQueueShards()
	shardLimit := limit / len(shardConfigs)
	for _, shardConfig := range shardConfigs {
		dbClient := client.NewClient(shardConfig.GetHost())
		var start []byte
		if startHeight != 0 {
			start = jutil.GetInt64DataBig(startHeight)
		}
		if err := dbClient.GetAll(ctx, db.TopicChainHeightBlock, start,
			client.NewOptionOrder(newest),
			client.NewOptionLimit(shardLimit),
		); err != nil {
			return nil, fmt.Errorf("error getting height blocks from queue client all; %w", err)
		}
		for i := range dbClient.Messages {
			var heightBlock = new(HeightBlock)
			db.Set(heightBlock, dbClient.Messages[i])
			heightBlocks = append(heightBlocks, heightBlock)
		}
	}
	sort.Slice(heightBlocks, func(i, j int) bool {
		if newest {
			return heightBlocks[i].Height > heightBlocks[j].Height
		} else {
			return heightBlocks[i].Height < heightBlocks[j].Height
		}
	})
	return heightBlocks, nil
}
