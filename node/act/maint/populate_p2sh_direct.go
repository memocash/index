package maint

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"sync"
	"time"
)

type PopulateP2shDirect struct {
	status   map[uint]*item.ProcessStatus
	hasError bool
	mu       sync.Mutex
	Checked  int64
	Saved    int64
}

func NewPopulateP2shDirect() *PopulateP2shDirect {
	return &PopulateP2shDirect{}
}

func (p *PopulateP2shDirect) SetShardStatus(shard uint32, status []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status[uint(shard)] = item.NewProcessStatus(uint(shard), item.ProcessStatusPopulateP2sh)
	p.status[uint(shard)].Status = status
}

func (p *PopulateP2shDirect) GetShardStatus(shard uint32) *item.ProcessStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	if shardStatus, ok := p.status[uint(shard)]; !ok {
		return nil
	} else {
		return shardStatus
	}
}

func (p *PopulateP2shDirect) SetHasError(hasError bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.hasError = hasError
}

func (p *PopulateP2shDirect) GetHasError() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hasError
}

func (p *PopulateP2shDirect) Populate(newRun bool) error {
	shardConfigs := config.GetQueueShards()
	if !newRun {
		for _, shardConfig := range shardConfigs {
			syncStatus, err := item.GetProcessStatus(uint(shardConfig.Shard), item.ProcessStatusPopulateP2sh)
			if err != nil && !client.IsEntryNotFoundError(err) {
				return jerr.Get("error getting sync status", err)
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
					jlog.Logf("Completed populating p2sh for shard: %d\n", config.Shard)
					wg.Done()
				} else if err != nil {
					errChan <- jerr.Getf(err, "error populating p2sh for shard: %d", config.Shard)
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
			return jerr.Get("error populating p2sh direct", err)
		case <-success:
			return nil
		case <-time.NewTimer(time.Second * 10).C:
			p.mu.Lock()
			jlog.Logf("Populating p2sh direct: %d checked, %d saved\n", p.Checked, p.Saved)
			p.mu.Unlock()
		}
	}
}

func (p *PopulateP2shDirect) populateShardSingle(shard uint32) (bool, error) {
	shardStatus := p.GetShardStatus(shard)
	if shardStatus == nil {
		shardStatus = item.NewProcessStatus(uint(shard), item.ProcessStatusPopulateP2sh)
	}
	txOutputs, err := chain.GetAllTxOutputs(shard, shardStatus.Status)
	if err != nil {
		return false, jerr.Getf(err, "error getting tx outputs for populate p2sh shard: %d", shard)
	}
	var txHeights = make(map[[32]byte]uint32)
	var objectsToSave []db.Object
	for _, txOutput := range txOutputs {
		uid := txOutput.GetUid()
		if jutil.ByteGT(uid, shardStatus.Status) {
			shardStatus.Status = uid
		}
		address, err := wallet.GetAddrFromLockScript(txOutput.LockScript)
		if err != nil || !address.IsP2SH() {
			continue
		}
		var height int64
		if _, ok := txHeights[txOutput.TxHash]; !ok {
			txBlocks, err := chain.GetSingleTxBlocks(txOutput.TxHash)
			if err != nil {
				return false, jerr.Getf(err, "error getting tx block for tx: %s", txOutput.TxHash)
			}
			if len(txBlocks) > 0 {
				blockHeight, err := chain.GetBlockHeight(txBlocks[0].BlockHash)
				if err != nil {
					return false, jerr.Getf(err, "error getting block height for block: %s", txBlocks[0].BlockHash)
				}
				height = blockHeight.Height
			} else {
				height = item.HeightMempool
			}
		}
		objectsToSave = append(objectsToSave, &addr.HeightOutput{
			Addr:   *address,
			Height: int32(height),
			TxHash: txOutput.TxHash,
			Index:  txOutput.Index,
			Value:  txOutput.Value,
		})
		spends, err := chain.GetOutputInput(memo.Out{
			TxHash: txOutput.TxHash[:],
			Index:  txOutput.Index,
		})
		if err != nil && !client.IsEntryNotFoundError(err) {
			return false, jerr.Getf(err, "error getting output input for tx: %s", txOutput.TxHash)
		}
		for _, spend := range spends {
			objectsToSave = append(objectsToSave, &addr.HeightInput{
				Addr:   *address,
				Height: int32(height),
				TxHash: spend.Hash,
				Index:  spend.Index,
			})
		}
	}
	if err := db.Save(objectsToSave); err != nil {
		return false, jerr.Get("error saving objects", err)
	}
	p.mu.Lock()
	p.Saved += int64(len(objectsToSave))
	p.Checked += int64(len(txOutputs))
	p.mu.Unlock()
	if err := shardStatus.Save(); err != nil {
		return false, jerr.Get("error saving process status", err)
	}
	p.SetShardStatus(shard, shardStatus.Status)
	return len(txOutputs) < client.HugeLimit, nil
}
