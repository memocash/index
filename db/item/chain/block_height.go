package chain

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"strings"
)

type BlockHeight struct {
	BlockHash [32]byte
	Height    int64
}

func (b *BlockHeight) GetTopic() string {
	return db.TopicChainBlockHeight
}

func (b *BlockHeight) GetShardSource() uint {
	return client.GenShardSource(b.BlockHash[:])
}

func (b *BlockHeight) GetUid() []byte {
	return jutil.CombineBytes(jutil.ByteReverse(b.BlockHash[:]), jutil.GetInt64DataBig(b.Height))
}

func (b *BlockHeight) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	copy(b.BlockHash[:], jutil.ByteReverse(uid[:32]))
	b.Height = jutil.GetInt64Big(uid[32:40])
}

func (b *BlockHeight) Serialize() []byte {
	return nil
}

func (b *BlockHeight) Deserialize([]byte) {}

func GetBlockHeight(blockHash [32]byte) (*BlockHeight, error) {
	shardConfig := config.GetShardConfig(client.GenShardSource32(blockHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetByPrefix(db.TopicChainBlockHeight, jutil.ByteReverse(blockHash[:])); err != nil {
		return nil, fmt.Errorf("error getting client message for block height; %w", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, fmt.Errorf("error no clients messages return for block height; %w", client.EntryNotFoundError)
	} else if len(dbClient.Messages) > 1 {
		var hashStrings = make([]string, len(dbClient.Messages))
		for i := range dbClient.Messages {
			hashStrings[i] = dbClient.Messages[i].UidHex()
		}
		return nil, fmt.Errorf("error more than 1 block height returned: %d (%s)",
			len(dbClient.Messages), strings.Join(hashStrings, ", "))
	}
	var blockHeight = new(BlockHeight)
	db.Set(blockHeight, dbClient.Messages[0])
	return blockHeight, nil
}

func GetBlockHeights(ctx context.Context, blockHashes [][32]byte) ([]*BlockHeight, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, blockHash := range blockHashes {
		shard := db.GetShardIdFromByte32(blockHash[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(blockHash[:]))
	}
	messages, err := db.GetByPrefixes(ctx, db.TopicChainBlockHeight, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting client message chain block heights; %w", err)
	}
	var blockHeights []*BlockHeight
	for _, msg := range messages {
		var blockHeight = new(BlockHeight)
		db.Set(blockHeight, msg)
		blockHeights = append(blockHeights, blockHeight)
	}
	return blockHeights, nil
}

func ListenBlockHeights(ctx context.Context) (chan *BlockHeight, error) {
	var chanBlockHeight = make(chan *BlockHeight)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(chanBlockHeight)
	})
	for _, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicChainBlockHeight, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting block height listen message chan; %w", err)
		}
		go func() {
			defer cancelCtx.Cancel()
			for msg := range chanMessage {
				var blockHeight = new(BlockHeight)
				db.Set(blockHeight, *msg)
				chanBlockHeight <- blockHeight
			}
		}()
	}
	return chanBlockHeight, nil
}
