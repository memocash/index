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
	"github.com/memocash/index/db/item/addr"
	item_cashtoken "github.com/memocash/index/db/item/cashtoken"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/tx/cashtoken"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
)

// PopulateCashToken backfills CashTokens for outputs saved before the live
// saver knew about them. It walks every chain_tx_output per shard and, for
// each output with a well-formed token prefix, writes the cash_token_output
// row and the addr_seen_tx row the address saver used to miss (the prefix
// hid the address). Progress is checkpointed per shard like populate-addr.
type PopulateCashToken struct {
	Context context.Context
	// LogInterval is how often progress is logged while running; zero means
	// the default of ten seconds.
	LogInterval time.Duration
	// checkpoints holds, per shard, the last chain_tx_output uid whose batch
	// was saved; the published value only moves after a successful save and
	// is only ever read or replaced under mu, never mutated in place.
	checkpoints map[uint32][]byte
	// err is the first worker error of the current run, set under mu.
	err     error
	mu      sync.Mutex
	Checked int64
	Tokens  int64
	Saved   int64
}

const populateCashTokenLogInterval = 10 * time.Second

func NewPopulateCashToken(ctx context.Context) *PopulateCashToken {
	return &PopulateCashToken{
		Context:     ctx,
		checkpoints: make(map[uint32][]byte),
	}
}

// setCheckpoint publishes a shard's checkpoint; the slice is copied so the
// caller's buffer can keep moving.
func (p *PopulateCashToken) setCheckpoint(shard uint32, uid []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.checkpoints[shard] = append([]byte{}, uid...)
}

// getCheckpoint returns a private copy of a shard's checkpoint (nil if none).
func (p *PopulateCashToken) getCheckpoint(shard uint32) []byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	if uid, ok := p.checkpoints[shard]; ok {
		return append([]byte{}, uid...)
	}
	return nil
}

// Populate runs one worker per shard until every shard is exhausted. The
// first worker error cancels the others; Populate waits for all of them to
// stop before returning that error, so no work outlives the call.
func (p *PopulateCashToken) Populate(newRun bool) error {
	shardConfigs := config.GetQueueShards()
	if !newRun {
		for _, shardConfig := range shardConfigs {
			syncStatus, err := item.GetProcessStatus(p.Context, uint(shardConfig.Shard), item.ProcessStatusPopulateCashToken)
			if err != nil && !client.IsMessageNotSetError(err) {
				return fmt.Errorf("error getting populate cash token status; %w", err)
			} else if syncStatus != nil {
				p.setCheckpoint(shardConfig.Shard, syncStatus.Status)
			}
		}
	}
	ctx, cancel := context.WithCancel(p.Context)
	defer cancel()
	var wg sync.WaitGroup
	for _, shardConfig := range shardConfigs {
		wg.Add(1)
		go func(shard uint32) {
			defer wg.Done()
			for {
				done, err := p.populateShardSingle(ctx, shard)
				if err != nil {
					// Recorded directly (not sent on a channel the main loop
					// selects on) so a failure can never be lost behind the
					// completion signal; the first one stops the other shards.
					p.recordError(fmt.Errorf("error populating cash tokens for shard: %d; %w", shard, err), cancel)
					return
				}
				if done {
					log.Printf("Completed populating cash tokens for shard: %d\n", shard)
					return
				}
				// Stop between pages once a sibling failed or the caller gave up.
				if ctx.Err() != nil {
					return
				}
			}
		}(shardConfig.Shard)
	}
	var workersDone = make(chan struct{})
	go func() {
		wg.Wait()
		close(workersDone)
	}()
	logInterval := p.LogInterval
	if logInterval <= 0 {
		logInterval = populateCashTokenLogInterval
	}
	ticker := time.NewTicker(logInterval)
	defer ticker.Stop()
	for {
		select {
		case <-workersDone:
			// Every worker has returned, so every error is recorded.
			if err := p.firstError(); err != nil {
				return err
			}
			// Workers stop quietly between pages when the caller's context
			// ends; that is not a completed backfill.
			return p.Context.Err()
		case <-ticker.C:
			p.logProgress()
		}
	}
}

// recordError keeps the first worker error and cancels the sibling workers.
func (p *PopulateCashToken) recordError(err error, cancel context.CancelFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err == nil {
		p.err = err
		cancel()
	}
}

func (p *PopulateCashToken) firstError() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

func (p *PopulateCashToken) logProgress() {
	p.mu.Lock()
	defer p.mu.Unlock()
	log.Printf("Populating cash tokens: %d outputs checked, %d token outputs, %d rows saved\n",
		p.Checked, p.Tokens, p.Saved)
	for shard, uid := range p.checkpoints {
		log.Printf("Shard %d status: %x\n", shard, uid)
	}
}

// populateShardSingle processes one page of a shard's outputs starting after
// its published checkpoint and reports whether the shard is exhausted. The
// checkpoint is advanced locally and published only once the page's rows and
// the process status row are saved.
func (p *PopulateCashToken) populateShardSingle(ctx context.Context, shard uint32) (bool, error) {
	checkpoint := p.getCheckpoint(shard)
	txOutputs, err := chain.GetAllTxOutputs(ctx, shard, checkpoint)
	if err != nil {
		return false, fmt.Errorf("error getting tx outputs for populate cash token shard: %d; %w", shard, err)
	}
	var objects []db.Object
	var seenTxAddrs = make(map[[32]byte][]wallet.Addr)
	var tokens int
	for _, txOutput := range txOutputs {
		uid := txOutput.GetUid()
		if jutil.ByteGT(uid, checkpoint) {
			checkpoint = uid
		}
		if !cashtoken.HasPrefix(txOutput.LockScript) {
			continue
		}
		token, _, err := cashtoken.Parse(txOutput.LockScript)
		if err != nil {
			log.Printf("error parsing cash token prefix %x:%d; %v\n", jutil.ByteReverse(txOutput.TxHash[:]), txOutput.Index, err)
			continue
		}
		tokens++
		objects = append(objects, item_cashtoken.NewOutput(txOutput.TxHash, txOutput.Index, *token))
		if address, err := wallet.GetAddrFromLockScript(txOutput.LockScript); err == nil {
			seenTxAddrs[txOutput.TxHash] = append(seenTxAddrs[txOutput.TxHash], *address)
		}
	}
	seenObjects, err := p.getSeenTxObjects(ctx, seenTxAddrs)
	if err != nil {
		return false, err
	}
	objects = append(objects, seenObjects...)
	if err := db.Save(objects); err != nil {
		return false, fmt.Errorf("error saving objects for populate cash token; %w", err)
	}
	shardStatus := item.NewProcessStatus(uint(shard), item.ProcessStatusPopulateCashToken)
	shardStatus.Status = checkpoint
	if err := shardStatus.Save(); err != nil {
		return false, fmt.Errorf("error saving populate cash token process status; %w", err)
	}
	p.setCheckpoint(shard, checkpoint)
	p.mu.Lock()
	p.Checked += int64(len(txOutputs))
	p.Tokens += int64(tokens)
	p.Saved += int64(len(objects))
	p.mu.Unlock()
	return len(txOutputs) < client.HugeLimit, nil
}

// getSeenTxObjects builds the addr_seen_tx rows for token outputs, dated by
// the tx's own first-seen time so address histories stay in order (falling
// back to now for a tx with no seen row).
func (p *PopulateCashToken) getSeenTxObjects(ctx context.Context, seenTxAddrs map[[32]byte][]wallet.Addr) ([]db.Object, error) {
	if len(seenTxAddrs) == 0 {
		return nil, nil
	}
	var txHashes = make([][32]byte, 0, len(seenTxAddrs))
	for txHash := range seenTxAddrs {
		txHashes = append(txHashes, txHash)
	}
	txSeens, err := chain.GetTxSeens(ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting tx seens for populate cash token; %w", err)
	}
	var seenTimes = make(map[[32]byte]time.Time, len(txSeens))
	for _, txSeen := range txSeens {
		if existing, ok := seenTimes[txSeen.TxHash]; !ok || txSeen.Timestamp.Before(existing) {
			seenTimes[txSeen.TxHash] = txSeen.Timestamp
		}
	}
	now := time.Now()
	var objects []db.Object
	var written = make(map[[57]byte]bool)
	for txHash, addrs := range seenTxAddrs {
		seen, ok := seenTimes[txHash]
		if !ok {
			seen = now
		}
		for _, address := range addrs {
			id := getAddrTxHashId(address, txHash)
			if written[id] {
				continue
			}
			written[id] = true
			objects = append(objects, &addr.SeenTx{Addr: address, TxHash: txHash, Seen: seen})
		}
	}
	return objects, nil
}
