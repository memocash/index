package maint

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type PopulateOutputInputSingle struct {
	Context      context.Context
	status       map[uint]*item.ProcessStatus
	hasError     bool
	mu           sync.Mutex
	Checked      int64
	Saved        int64
	DoubleSpends int64
}

func NewPopulateOutputInputSingle(ctx context.Context) *PopulateOutputInputSingle {
	return &PopulateOutputInputSingle{
		Context: ctx,
		status:  make(map[uint]*item.ProcessStatus),
	}
}

func (p *PopulateOutputInputSingle) SetShardStatus(shard uint32, status []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status[uint(shard)] = item.NewProcessStatus(uint(shard), item.ProcessStatusPopulateOutputInputSingle)
	p.status[uint(shard)].Status = status
}

func (p *PopulateOutputInputSingle) GetShardStatus(shard uint32) *item.ProcessStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	if shardStatus, ok := p.status[uint(shard)]; !ok {
		return nil
	} else {
		return shardStatus
	}
}

func (p *PopulateOutputInputSingle) SetHasError(hasError bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.hasError = hasError
}

func (p *PopulateOutputInputSingle) GetHasError() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hasError
}

func (p *PopulateOutputInputSingle) Populate(newRun bool) error {
	shardConfigs := config.GetQueueShards()
	if !newRun {
		for _, shardConfig := range shardConfigs {
			syncStatus, err := item.GetProcessStatus(p.Context, uint(shardConfig.Shard), item.ProcessStatusPopulateOutputInputSingle)
			if err != nil && !client.IsMessageNotSetError(err) {
				return fmt.Errorf("error getting sync status; %w", err)
			} else if syncStatus != nil {
				p.SetShardStatus(shardConfig.Shard, syncStatus.Status)
			}
		}
	}
	var wg sync.WaitGroup
	wg.Add(len(shardConfigs))
	var errChan = make(chan error)
	for _, shardConfig := range shardConfigs {
		go func(config config.Shard) {
			for {
				if done, err := p.populateShardSingle(config.Shard); done {
					log.Printf("Completed populating output input single for shard: %d\n", config.Shard)
					wg.Done()
				} else if err != nil {
					errChan <- fmt.Errorf("error populating output input single for shard: %d; %w", config.Shard, err)
				} else {
					continue
				}
				return
			}
		}(shardConfig)
	}
	var success = make(chan bool)
	go func() {
		wg.Wait()
		success <- true
	}()
	for {
		select {
		case err := <-errChan:
			p.SetHasError(true)
			return fmt.Errorf("error populating output input single; %w", err)
		case <-success:
			return nil
		case <-time.NewTimer(time.Second * 10).C:
			p.mu.Lock()
			log.Printf("Populating output input single: %d checked, %d saved, %d double spends\n",
				p.Checked, p.Saved, p.DoubleSpends)
			for shard, status := range p.status {
				log.Printf("Shard %d status: %x\n", shard, status.Status)
			}
			p.mu.Unlock()
		}
	}
}

func (p *PopulateOutputInputSingle) populateShardSingle(shard uint32) (bool, error) {
	shardStatus := p.GetShardStatus(shard)
	if shardStatus == nil {
		shardStatus = item.NewProcessStatus(uint(shard), item.ProcessStatusPopulateOutputInputSingle)
	}
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetByPrefix(p.Context, db.TopicChainOutputInput, client.Prefix{
		Start: shardStatus.Status,
	}, client.OptionHugeLimit()); err != nil {
		return false, fmt.Errorf("error getting output inputs for shard: %d; %w", shard, err)
	}
	var checked = len(dbClient.Messages)
	isFinalBatch := checked < client.HugeLimit
	// Group output inputs by output (PrevHash + PrevIndex = first 36 bytes of UID)
	type outputKey [36]byte
	grouped := make(map[outputKey][]*chain.OutputInput)
	var lastOutputKey outputKey
	for _, msg := range dbClient.Messages {
		var oi = new(chain.OutputInput)
		db.Set(oi, msg)
		uid := oi.GetUid()
		if jutil.ByteGT(uid, shardStatus.Status) {
			shardStatus.Status = uid
		}
		if memo.IsCoinbase(oi.PrevHash[:], oi.PrevIndex) {
			continue
		}
		copy(lastOutputKey[:], msg.Uid[:36])
		grouped[lastOutputKey] = append(grouped[lastOutputKey], oi)
	}
	// Exclude last output group if not the final batch, since it may be split across batches
	if !isFinalBatch {
		delete(grouped, lastOutputKey)
	}
	var objectsToSave []db.Object
	var doubleSpends int64
	for _, ois := range grouped {
		var winner *chain.OutputInput
		if len(ois) == 1 {
			winner = ois[0]
		} else {
			doubleSpends++
			resolved, err := p.resolveDoubleSpend(ois)
			if err != nil {
				log.Printf("Error resolving double spend for output %x:%d: %v\n",
					ois[0].PrevHash, ois[0].PrevIndex, err)
				continue
			}
			if resolved == nil {
				continue
			}
			winner = resolved
		}
		objectsToSave = append(objectsToSave, &chain.OutputInputSingle{
			PrevHash:  winner.PrevHash,
			PrevIndex: winner.PrevIndex,
			Hash:      winner.Hash,
			Index:     winner.Index,
		})
	}
	if err := db.Save(objectsToSave); err != nil {
		return false, fmt.Errorf("error saving output input singles; %w", err)
	}
	p.mu.Lock()
	p.Saved += int64(len(objectsToSave))
	p.Checked += int64(checked)
	p.DoubleSpends += doubleSpends
	p.mu.Unlock()
	if err := shardStatus.Save(); err != nil {
		return false, fmt.Errorf("error saving process status; %w", err)
	}
	p.SetShardStatus(shard, shardStatus.Status)
	return isFinalBatch, nil
}

// resolveDoubleSpend determines which of the spending txs is actually mined in a block.
// It looks up TxBlock for each spending tx and checks which block is on the main chain.
func (p *PopulateOutputInputSingle) resolveDoubleSpend(ois []*chain.OutputInput) (*chain.OutputInput, error) {
	var txHashes [][32]byte
	for _, oi := range ois {
		txHashes = append(txHashes, oi.Hash)
	}
	txBlocks, err := chain.GetTxBlocks(p.Context, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting tx blocks; %w", err)
	}
	// Map spending tx hash to its block hashes
	txBlockMap := make(map[[32]byte][][32]byte)
	var blockHashes [][32]byte
	for _, txBlock := range txBlocks {
		txBlockMap[txBlock.TxHash] = append(txBlockMap[txBlock.TxHash], txBlock.BlockHash)
		blockHashes = append(blockHashes, txBlock.BlockHash)
	}
	if len(blockHashes) == 0 {
		return nil, nil
	}
	// Get heights for all blocks to find which are on the main chain
	blockHeights, err := chain.GetBlockHeights(p.Context, blockHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting block heights; %w", err)
	}
	// Build set of main chain block hashes by checking HeightBlock at each height
	mainChainBlocks := make(map[[32]byte]bool)
	for _, bh := range blockHeights {
		heightBlocks, err := chain.GetHeightBlock(p.Context, bh.Height)
		if err != nil {
			return nil, fmt.Errorf("error getting height block for height %d; %w", bh.Height, err)
		}
		if len(heightBlocks) == 1 {
			mainChainBlocks[heightBlocks[0].BlockHash] = true
		} else if len(heightBlocks) > 1 {
			// Multiple blocks at same height - use the first one (primary)
			mainChainBlocks[heightBlocks[0].BlockHash] = true
		}
	}
	// Find the spending tx that is in a main chain block
	for _, oi := range ois {
		blocks, ok := txBlockMap[oi.Hash]
		if !ok {
			continue
		}
		for _, blockHash := range blocks {
			if mainChainBlocks[blockHash] {
				return oi, nil
			}
		}
	}
	// If none found on main chain, check if any tx has a block at all (prefer the one with a block)
	for _, oi := range ois {
		if _, ok := txBlockMap[oi.Hash]; ok {
			return oi, nil
		}
	}
	return nil, nil
}

