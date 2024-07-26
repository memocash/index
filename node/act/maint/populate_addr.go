package maint

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"log"
	"sync"
	"time"
)

type PopulateAddr struct {
	Context  context.Context
	status   map[uint]*item.ProcessStatus
	hasError bool
	mu       sync.Mutex
	Checked  int64
	Saved    int64
	Inputs   bool
}

func NewPopulateAddr(ctx context.Context, inputs bool) *PopulateAddr {
	return &PopulateAddr{
		Context: ctx,
		status:  make(map[uint]*item.ProcessStatus),
		Inputs:  inputs,
	}
}

func (p *PopulateAddr) GetStatusName() string {
	if p.Inputs {
		return item.ProcessStatusPopulateAddrInputs
	} else {
		return item.ProcessStatusPopulateAddr
	}
}

func (p *PopulateAddr) SetShardStatus(shard uint32, status []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status[uint(shard)] = item.NewProcessStatus(uint(shard), p.GetStatusName())
	p.status[uint(shard)].Status = status
}

func (p *PopulateAddr) GetShardStatus(shard uint32) *item.ProcessStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	if shardStatus, ok := p.status[uint(shard)]; !ok {
		return nil
	} else {
		return shardStatus
	}
}

func (p *PopulateAddr) SetHasError(hasError bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.hasError = hasError
}

func (p *PopulateAddr) GetHasError() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hasError
}

func (p *PopulateAddr) Populate(newRun bool) error {
	shardConfigs := config.GetQueueShards()
	if !newRun {
		for _, shardConfig := range shardConfigs {
			syncStatus, err := item.GetProcessStatus(p.Context, uint(shardConfig.Shard), p.GetStatusName())
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
					log.Printf("Completed populating addr for shard: %d\n", config.Shard)
					wg.Done()
				} else if err != nil {
					errChan <- fmt.Errorf("error populating addr for shard: %d; %w", config.Shard, err)
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
			return fmt.Errorf("error populating addr direct; %w", err)
		case <-success:
			return nil
		case <-time.NewTimer(time.Second * 10).C:
			p.mu.Lock()
			log.Printf("Populating addr direct: %d checked, %d saved\n", p.Checked, p.Saved)
			for shard, status := range p.status {
				log.Printf("Shard %d status: %x\n", shard, status.Status)
			}
			p.mu.Unlock()
		}
	}
}

func (p *PopulateAddr) populateShardSingle(shard uint32) (bool, error) {
	shardStatus := p.GetShardStatus(shard)
	if shardStatus == nil {
		shardStatus = item.NewProcessStatus(uint(shard), p.GetStatusName())
	}
	seenTime := time.Now()
	var objMap = make(map[[57]byte]*addr.SeenTx)
	var checked int
	if p.Inputs {
		txInputs, err := chain.GetAllTxInputs(p.Context, shard, shardStatus.Status)
		if err != nil {
			return false, fmt.Errorf("error getting tx outputs for populate addr shard: %d; %w", shard, err)
		}
		for _, txInput := range txInputs {
			uid := txInput.GetUid()
			if jutil.ByteGT(uid, shardStatus.Status) {
				shardStatus.Status = uid
			}
			address, err := wallet.GetAddrFromUnlockScript(txInput.UnlockScript)
			if err != nil {
				continue
			}
			objMap[getAddrTxHashId(*address, txInput.TxHash)] = &addr.SeenTx{
				Addr:   *address,
				TxHash: txInput.TxHash,
				Seen:   seenTime,
			}
		}
		checked = len(txInputs)
	} else {
		txOutputs, err := chain.GetAllTxOutputs(shard, shardStatus.Status)
		if err != nil {
			return false, fmt.Errorf("error getting tx outputs for populate addr shard: %d; %w", shard, err)
		}
		for _, txOutput := range txOutputs {
			uid := txOutput.GetUid()
			if jutil.ByteGT(uid, shardStatus.Status) {
				shardStatus.Status = uid
			}
			address, err := wallet.GetAddrFromLockScript(txOutput.LockScript)
			if err != nil {
				continue
			}
			objMap[getAddrTxHashId(*address, txOutput.TxHash)] = &addr.SeenTx{
				Addr:   *address,
				TxHash: txOutput.TxHash,
				Seen:   seenTime,
			}
		}
		checked = len(txOutputs)
	}
	var objectsToSave = make([]db.Object, 0, len(objMap))
	for _, obj := range objMap {
		objectsToSave = append(objectsToSave, obj)
	}
	if err := db.Save(objectsToSave); err != nil {
		return false, fmt.Errorf("error saving objects for addr populate single; %w", err)
	}
	p.mu.Lock()
	p.Saved += int64(len(objectsToSave))
	p.Checked += int64(checked)
	p.mu.Unlock()
	if err := shardStatus.Save(); err != nil {
		return false, fmt.Errorf("error saving process status; %w", err)
	}
	p.SetShardStatus(shard, shardStatus.Status)
	return checked < client.HugeLimit, nil
}
