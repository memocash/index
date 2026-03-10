package maint

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type ListHeightDuplicates struct {
	CheckDoubleSpends bool
	Total             int
	DoubleSpends      int
}

func NewListHeightDuplicates(checkDoubleSpends bool) *ListHeightDuplicates {
	return &ListHeightDuplicates{
		CheckDoubleSpends: checkDoubleSpends,
	}
}

func (l *ListHeightDuplicates) List(ctx context.Context) error {
	for i, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		for {
			if err := dbClient.GetByPrefix(ctx, db.TopicChainHeightDuplicate, client.Prefix{
				Start: startUid,
			}, client.OptionHugeLimit()); err != nil {
				return fmt.Errorf("error getting height duplicates for shard %d; %w", i, err)
			}
			for _, msg := range dbClient.Messages {
				var heightDuplicate chain.HeightDuplicate
				heightDuplicate.SetUid(msg.Uid)
				log.Printf("Height: %d, Block: %s\n", heightDuplicate.Height, chainhash.Hash(heightDuplicate.BlockHash).String())
				l.Total++
				if l.CheckDoubleSpends {
					if err := l.checkBlockDoubleSpends(ctx, heightDuplicate.Height, heightDuplicate.BlockHash); err != nil {
						return fmt.Errorf("error checking double spends for height %d; %w", heightDuplicate.Height, err)
					}
				}
			}
			if len(dbClient.Messages) < client.HugeLimit {
				break
			}
			startUid = dbClient.Messages[len(dbClient.Messages)-1].Uid
		}
	}
	return nil
}

func (l *ListHeightDuplicates) checkBlockDoubleSpends(ctx context.Context, height int64, blockHash [32]byte) error {
	var allTxHashes [][32]byte
	var startIndex uint32
	for {
		blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{
			Context:    ctx,
			BlockHash:  blockHash,
			StartIndex: startIndex,
		})
		if err != nil {
			return fmt.Errorf("error getting block txs; %w", err)
		}
		for _, blockTx := range blockTxs {
			allTxHashes = append(allTxHashes, blockTx.TxHash)
		}
		if len(blockTxs) < client.LargeLimit {
			break
		}
		startIndex = blockTxs[len(blockTxs)-1].Index + 1
	}
	const batchSize = client.LargeLimit
	var outs []memo.Out
	for i := 0; i < len(allTxHashes); i += batchSize {
		end := i + batchSize
		if end > len(allTxHashes) {
			end = len(allTxHashes)
		}
		txInputs, err := chain.GetTxInputsByHashes(ctx, allTxHashes[i:end])
		if err != nil {
			return fmt.Errorf("error getting tx inputs; %w", err)
		}
		for _, txInput := range txInputs {
			if memo.IsCoinbase(txInput.PrevHash[:], txInput.PrevIndex) {
				continue
			}
			outs = append(outs, memo.Out{
				TxHash: txInput.PrevHash[:],
				Index:  txInput.PrevIndex,
			})
		}
	}
	if len(outs) == 0 {
		return nil
	}
	type outKey struct {
		Hash  [32]byte
		Index uint32
	}
	spendCounts := make(map[outKey]int)
	for i := 0; i < len(outs); i += batchSize {
		end := i + batchSize
		if end > len(outs) {
			end = len(outs)
		}
		outputInputs, err := chain.GetOutputInputs(ctx, outs[i:end])
		if err != nil {
			return fmt.Errorf("error getting output inputs; %w", err)
		}
		for _, oi := range outputInputs {
			key := outKey{Hash: oi.PrevHash, Index: oi.PrevIndex}
			spendCounts[key]++
		}
	}
	for key, count := range spendCounts {
		if count > 1 {
			l.DoubleSpends++
			log.Printf("  Double spend at height %d: output %s:%d spent %d times\n",
				height, chainhash.Hash(key.Hash).String(), key.Index, count)
		}
	}
	return nil
}
