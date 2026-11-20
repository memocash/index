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
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
)

// PopulateP2sh32 backfills what the address saver missed before it knew
// P2SH32. In outputs mode it walks every chain_tx_output per shard and
// writes the addr_seen_tx row for each P2SH32 lock script, dated by the tx's
// own first-seen time. In links mode it walks every chain_tx_input and
// writes the P2SH20/P2SH32 link for each P2SH-shaped unlock script. The
// input-side seen rows need no backfill: spends were always recorded under
// the P2SH20 alias, which is exactly where the link points. Progress is
// checkpointed per shard like populate-cash-token.
type PopulateP2sh32 struct {
	Context context.Context
	Links   bool
	// LogInterval is how often progress is logged while running; zero means
	// the default of ten seconds.
	LogInterval time.Duration
	// checkpoints holds, per shard, the last uid whose batch was saved; the
	// published value only moves after a successful save and is only ever
	// read or replaced under mu, never mutated in place.
	checkpoints map[uint32][]byte
	// err is the first worker error of the current run, set under mu.
	err     error
	mu      sync.Mutex
	Checked int64
	Found   int64
	Saved   int64
}

const populateP2sh32LogInterval = 10 * time.Second

func NewPopulateP2sh32(ctx context.Context, links bool) *PopulateP2sh32 {
	return &PopulateP2sh32{
		Context:     ctx,
		Links:       links,
		checkpoints: make(map[uint32][]byte),
	}
}

func (p *PopulateP2sh32) statusName() string {
	if p.Links {
		return item.ProcessStatusPopulateP2shLinks
	}
	return item.ProcessStatusPopulateP2sh32
}

func (p *PopulateP2sh32) what() string {
	if p.Links {
		return "p2sh links"
	}
	return "p2sh32 outputs"
}

// setCheckpoint publishes a shard's checkpoint; the slice is copied so the
// caller's buffer can keep moving.
func (p *PopulateP2sh32) setCheckpoint(shard uint32, uid []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.checkpoints[shard] = append([]byte{}, uid...)
}

// getCheckpoint returns a private copy of a shard's checkpoint (nil if none).
func (p *PopulateP2sh32) getCheckpoint(shard uint32) []byte {
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
func (p *PopulateP2sh32) Populate(newRun bool) error {
	shardConfigs := config.GetQueueShards()
	if !newRun {
		for _, shardConfig := range shardConfigs {
			syncStatus, err := item.GetProcessStatus(p.Context, uint(shardConfig.Shard), p.statusName())
			if err != nil && !client.IsMessageNotSetError(err) {
				return fmt.Errorf("error getting populate %s status; %w", p.what(), err)
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
					p.recordError(fmt.Errorf("error populating %s for shard: %d; %w", p.what(), shard, err), cancel)
					return
				}
				if done {
					log.Printf("Completed populating %s for shard: %d\n", p.what(), shard)
					return
				}
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
		logInterval = populateP2sh32LogInterval
	}
	ticker := time.NewTicker(logInterval)
	defer ticker.Stop()
	for {
		select {
		case <-workersDone:
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
func (p *PopulateP2sh32) recordError(err error, cancel context.CancelFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err == nil {
		p.err = err
		cancel()
	}
}

func (p *PopulateP2sh32) firstError() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

func (p *PopulateP2sh32) logProgress() {
	p.mu.Lock()
	defer p.mu.Unlock()
	log.Printf("Populating %s: %d checked, %d found, %d rows saved\n", p.what(), p.Checked, p.Found, p.Saved)
	for shard, uid := range p.checkpoints {
		log.Printf("Shard %d status: %x\n", shard, uid)
	}
}

// populateShardSingle processes one page of a shard starting after its
// published checkpoint and reports whether the shard is exhausted. The
// checkpoint is advanced locally and published only once the page's rows and
// the process status row are saved.
func (p *PopulateP2sh32) populateShardSingle(ctx context.Context, shard uint32) (bool, error) {
	checkpoint := p.getCheckpoint(shard)
	var objects []db.Object
	var checked, found int
	var err error
	if p.Links {
		objects, checked, found, checkpoint, err = p.linkObjects(ctx, shard, checkpoint)
	} else {
		objects, checked, found, checkpoint, err = p.outputObjects(ctx, shard, checkpoint)
	}
	if err != nil {
		return false, err
	}
	if err := db.Save(objects); err != nil {
		return false, fmt.Errorf("error saving objects for populate %s; %w", p.what(), err)
	}
	shardStatus := item.NewProcessStatus(uint(shard), p.statusName())
	shardStatus.Status = checkpoint
	if err := shardStatus.Save(); err != nil {
		return false, fmt.Errorf("error saving populate %s process status; %w", p.what(), err)
	}
	p.setCheckpoint(shard, checkpoint)
	p.mu.Lock()
	p.Checked += int64(checked)
	p.Found += int64(found)
	p.Saved += int64(len(objects))
	p.mu.Unlock()
	return checked < client.HugeLimit, nil
}

// linkObjects walks a page of inputs and returns a link pair per distinct
// redeem script found in a P2SH-shaped unlock script.
func (p *PopulateP2sh32) linkObjects(ctx context.Context, shard uint32, checkpoint []byte) ([]db.Object, int, int, []byte, error) {
	txInputs, err := chain.GetAllTxInputs(ctx, shard, checkpoint)
	if err != nil {
		return nil, 0, 0, nil, fmt.Errorf("error getting tx inputs for populate p2sh links shard: %d; %w", shard, err)
	}
	var links = make(map[wallet.Addr32]wallet.Addr)
	var found int
	for _, txInput := range txInputs {
		uid := txInput.GetUid()
		if jutil.ByteGT(uid, checkpoint) {
			checkpoint = uid
		}
		if _, err := wallet.GetP2pkhAddrFromUnlockScript(txInput.UnlockScript); err == nil {
			continue
		}
		redeemScript, err := wallet.GetP2shRedeemScriptFromUnlockScript(txInput.UnlockScript)
		if err != nil {
			continue
		}
		found++
		p2shAddr, p2sh32Addr := wallet.GetP2shAddrsFromRedeemScript(redeemScript)
		links[p2sh32Addr] = p2shAddr
	}
	var objects = make([]db.Object, 0, 2*len(links))
	for p2sh32Addr, p2shAddr := range links {
		objects = append(objects, addr.NewP2shLinks(p2shAddr, p2sh32Addr)...)
	}
	return objects, len(txInputs), found, checkpoint, nil
}

// outputObjects walks a page of outputs and returns a seen-tx row per
// P2SH32 output, dated by the tx's own first-seen time so address histories
// stay in order (falling back to now for a tx with no seen row).
func (p *PopulateP2sh32) outputObjects(ctx context.Context, shard uint32, checkpoint []byte) ([]db.Object, int, int, []byte, error) {
	txOutputs, err := chain.GetAllTxOutputs(ctx, shard, checkpoint)
	if err != nil {
		return nil, 0, 0, nil, fmt.Errorf("error getting tx outputs for populate p2sh32 shard: %d; %w", shard, err)
	}
	var seenTxAddrs = make(map[[32]byte][]wallet.LockAddr)
	var found int
	for _, txOutput := range txOutputs {
		uid := txOutput.GetUid()
		if jutil.ByteGT(uid, checkpoint) {
			checkpoint = uid
		}
		address, err := wallet.GetLockAddrFromLockScript(txOutput.LockScript)
		if err != nil || !address.IsP2SH32() {
			continue
		}
		found++
		seenTxAddrs[txOutput.TxHash] = append(seenTxAddrs[txOutput.TxHash], address)
	}
	objects, err := getSeenTxObjects(ctx, seenTxAddrs)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	return objects, len(txOutputs), found, checkpoint, nil
}

// getSeenTxObjects builds the addr_seen_tx rows for the given txs' addresses,
// dated by each tx's first-seen time.
func getSeenTxObjects(ctx context.Context, seenTxAddrs map[[32]byte][]wallet.LockAddr) ([]db.Object, error) {
	if len(seenTxAddrs) == 0 {
		return nil, nil
	}
	var txHashes = make([][32]byte, 0, len(seenTxAddrs))
	for txHash := range seenTxAddrs {
		txHashes = append(txHashes, txHash)
	}
	txSeens, err := chain.GetTxSeens(ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting tx seens for populate; %w", err)
	}
	var seenTimes = make(map[[32]byte]time.Time, len(txSeens))
	for _, txSeen := range txSeens {
		if existing, ok := seenTimes[txSeen.TxHash]; !ok || txSeen.Timestamp.Before(existing) {
			seenTimes[txSeen.TxHash] = txSeen.Timestamp
		}
	}
	now := time.Now()
	var objects []db.Object
	var written = make(map[addrTxKey]bool)
	for txHash, addrs := range seenTxAddrs {
		seen, ok := seenTimes[txHash]
		if !ok {
			seen = now
		}
		for _, address := range addrs {
			key := getAddrTxKey(address, txHash)
			if written[key] {
				continue
			}
			written[key] = true
			objects = append(objects, &addr.SeenTx{Addr: address, TxHash: txHash, Seen: seen})
		}
	}
	return objects, nil
}
