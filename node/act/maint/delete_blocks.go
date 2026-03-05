package maint

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type DeleteBlocks struct {
	Ctx        context.Context
	Verbose    bool
	DryRun     bool
	Start      int64
	Blocks     int
	TxLinks    int
	Duplicates int
}

func NewDeleteBlocks(ctx context.Context, start int64, verbose bool, dryRun bool) *DeleteBlocks {
	return &DeleteBlocks{
		Ctx:     ctx,
		Start:   start,
		Verbose: verbose,
		DryRun:  dryRun,
	}
}

func (d *DeleteBlocks) Delete() error {
	var nextHeight = d.Start
	for {
		heightBlocks, err := chain.GetHeightBlocksAllLimit(d.Ctx, nextHeight, client.HugeLimit, false)
		if err != nil {
			return fmt.Errorf("error getting height blocks; %w", err)
		}
		if len(heightBlocks) == 0 {
			break
		}
		for _, hb := range heightBlocks {
			if err := d.deleteBlock(hb); err != nil {
				return fmt.Errorf("error deleting block at height %d; %w", hb.Height, err)
			}
			nextHeight = hb.Height + 1
		}
	}
	if err := d.deleteHeightDuplicates(); err != nil {
		return fmt.Errorf("error deleting height duplicates; %w", err)
	}
	return nil
}

func (d *DeleteBlocks) deleteBlock(hb *chain.HeightBlock) error {
	blockHash := chainhash.Hash(hb.BlockHash)
	txLinks, err := d.getBlockTxObjects(hb.BlockHash)
	if err != nil {
		return fmt.Errorf("error getting block tx objects; %w", err)
	}
	txLinkCount := len(txLinks) / 2
	if d.Verbose || d.DryRun {
		log.Printf("Height %d: %s (%d txs)", hb.Height, blockHash, txLinkCount)
	}
	d.TxLinks += txLinkCount
	d.Blocks++
	if d.DryRun {
		return nil
	}
	var objects []db.Object
	objects = append(objects,
		&chain.HeightBlock{Height: hb.Height, BlockHash: hb.BlockHash},
		&chain.BlockHeight{Height: hb.Height, BlockHash: hb.BlockHash},
		&chain.Block{Hash: hb.BlockHash},
		&chain.BlockInfo{BlockHash: hb.BlockHash},
	)
	objects = append(objects, txLinks...)
	if err := db.Remove(objects); err != nil {
		return fmt.Errorf("error removing block objects; %w", err)
	}
	return nil
}

func (d *DeleteBlocks) getBlockTxObjects(blockHash [32]byte) ([]db.Object, error) {
	var objects []db.Object
	var startIndex uint32
	for {
		blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{
			Context:    d.Ctx,
			BlockHash:  blockHash,
			StartIndex: startIndex,
			Limit:      client.LargeLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("error getting block txs; %w", err)
		}
		for _, bt := range blockTxs {
			objects = append(objects,
				&chain.BlockTx{BlockHash: bt.BlockHash, Index: bt.Index, TxHash: bt.TxHash},
				&chain.TxBlock{TxHash: bt.TxHash, BlockHash: bt.BlockHash},
			)
			startIndex = bt.Index + 1
		}
		if len(blockTxs) < int(client.LargeLimit) {
			break
		}
	}
	return objects, nil
}

func (d *DeleteBlocks) deleteHeightDuplicates() error {
	for i, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		if d.Start > 0 {
			startUid = jutil.GetInt64DataBig(d.Start)
		}
		for {
			if err := dbClient.GetByPrefix(d.Ctx, db.TopicChainHeightDuplicate, client.Prefix{
				Start: startUid,
			}, client.OptionHugeLimit()); err != nil {
				return fmt.Errorf("error getting height duplicates for shard %d; %w", i, err)
			}
			if len(dbClient.Messages) == 0 {
				break
			}
			var objects []db.Object
			for _, msg := range dbClient.Messages {
				var hd chain.HeightDuplicate
				hd.SetUid(msg.Uid)
				if hd.Height >= d.Start {
					if d.Verbose || d.DryRun {
						log.Printf("Height duplicate at height %d: %s",
							hd.Height, chainhash.Hash(hd.BlockHash))
					}
					d.Duplicates++
					objects = append(objects, &chain.HeightDuplicate{
						Height:    hd.Height,
						BlockHash: hd.BlockHash,
					})
				}
			}
			if len(objects) > 0 && !d.DryRun {
				if err := db.Remove(objects); err != nil {
					return fmt.Errorf("error removing height duplicates; %w", err)
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
