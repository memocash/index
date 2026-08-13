package maint

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/node/act/slp_validate"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
	"github.com/memocash/index/ref/config"
)

const (
	slpBackfillScanLimit        = 100 * 1000 // matches per filtered-scan page
	slpBackfillBatchSize        = 10000      // hashes per batched lookup
	slpBackfillChunkTarget      = 10000      // txs per validation chunk (never splits a block)
	slpBackfillLogChunkInterval = 25
	// A single scan page runs server-side until it finds a page of matches or
	// exhausts the shard, so a cold-cache call over a sparse stretch can far
	// outrun the client's 60s default RPC deadline, and one timeout aborts
	// the whole backfill
	slpBackfillScanTimeout = 10 * time.Minute
)

// SlpValidityBackfill validates every undecided SLP tx exactly once by
// visiting them in block order. Server-side filtered scans find the SLP
// candidates without shipping the whole chain_tx_output topic over gRPC;
// parents are always in earlier blocks or the same block, so validating
// height-ordered chunks (topologically sorted within a chunk — CTOR blocks
// order txs by txid, not topologically) decides each tx in a single
// validation call instead of the cascade's one-generation-per-round walk.
// Mempool txs are cascaded at the end. Verdicts are final, so re-running is
// idempotent and already-decided txs drop out at the collect phase.
type SlpValidityBackfill struct {
	Ctx     context.Context
	Verbose bool

	Candidates  int64 // distinct SLP vout-0 candidate txs found
	Decided     int64 // candidates that already had a verdict
	// SlpTxs counts undecided txs run through validation. COMMIT-type txs are
	// included but never receive a verdict (validation skips them — the
	// deferred COMMIT liveness gap), so they are recollected by every run and
	// SlpTxs can exceed Valid+Invalid+Pending
	SlpTxs int64
	Valid       int64
	Invalid     int64
	Pending     int64
	Missing     int64 // candidates whose chain rows are incomplete
	MempoolTail int64 // candidates with no mined block, cascaded at the end
}

func NewSlpValidityBackfill(ctx context.Context, verbose bool) *SlpValidityBackfill {
	return &SlpValidityBackfill{
		Ctx:     ctx,
		Verbose: verbose,
	}
}

func (b *SlpValidityBackfill) Run() error {
	undecided, err := b.collect()
	if err != nil {
		return fmt.Errorf("error collecting slp candidates for validity backfill; %w", err)
	}
	log.Printf("slp validity backfill collected: candidates %d, already decided %d, undecided %d\n",
		b.Candidates, b.Decided, len(undecided))
	ordered, mempool, err := b.order(undecided)
	if err != nil {
		return fmt.Errorf("error ordering slp candidates for validity backfill; %w", err)
	}
	b.MempoolTail = int64(len(mempool))
	chunks := chunkByBlock(ordered)
	log.Printf("slp validity backfill ordered: %d mined txs in %d chunks, %d mempool tail\n",
		len(ordered), len(chunks), len(mempool))
	validator := slp_validate.NewValidator()
	for i, chunk := range chunks {
		if err := b.processChunk(validator, chunk); err != nil {
			return fmt.Errorf("error processing slp validity backfill chunk %d; %w", i, err)
		}
		if (i+1)%slpBackfillLogChunkInterval == 0 || i+1 == len(chunks) {
			log.Printf("slp validity backfill chunk %d/%d (height %d): slp txs %d, valid %d, invalid %d, pending %d, missing %d\n",
				i+1, len(chunks), chunk[0].height, b.SlpTxs, b.Valid, b.Invalid, b.Pending, b.Missing)
		}
	}
	if err := b.tail(mempool); err != nil {
		return fmt.Errorf("error processing slp validity backfill mempool tail; %w", err)
	}
	log.Printf("slp validity backfill done. candidates: %d, decided: %d, slp txs: %d, "+
		"valid: %d, invalid: %d, pending: %d, missing: %d, mempool tail: %d\n",
		b.Candidates, b.Decided, b.SlpTxs, b.Valid, b.Invalid, b.Pending, b.Missing, b.MempoolTail)
	return nil
}

// collect finds every undecided SLP candidate: a server-side filtered scan
// over chain_tx_output per shard (uid suffix 00000000 = vout 0, value
// containing the SLP lokad), re-verified client-side with HasSlpLokad (a
// contains match can hit lokad bytes elsewhere in a script), then filtered
// against existing verdicts.
func (b *SlpValidityBackfill) collect() ([][32]byte, error) {
	var lock sync.Mutex
	var wg sync.WaitGroup
	var errs []error
	var undecided [][32]byte
	for _, shardConfig := range config.GetQueueShards() {
		wg.Add(1)
		go func(shard uint32) {
			defer wg.Done()
			shardUndecided, candidates, decided, err := b.collectShard(shard)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("error collecting slp candidates shard %d; %w", shard, err))
				return
			}
			undecided = append(undecided, shardUndecided...)
			b.Candidates += candidates
			b.Decided += decided
		}(shardConfig.Shard)
	}
	wg.Wait()
	if len(errs) > 0 {
		return nil, jerr.Combine(errs...)
	}
	return undecided, nil
}

func (b *SlpValidityBackfill) collectShard(shard uint32) ([][32]byte, int64, int64, error) {
	var candidates [][32]byte
	var start []byte
	for {
		messages, err := db.Search(b.Ctx, db.TopicChainTxOutput, shard, client.SearchPattern{
			Start: start,
			Uid:   client.NewPatternSuffix([]byte{0, 0, 0, 0}),
			Data:  client.NewPatternContains(memo.PrefixSlp),
		}, slpBackfillScanLimit, client.NewOptionTimeout(slpBackfillScanTimeout))
		if err != nil {
			return nil, 0, 0, fmt.Errorf("error scanning tx outputs for slp candidates; %w", err)
		}
		for _, msg := range messages {
			// uid = 32-byte reversed tx hash + 4-byte big-endian index; the
			// suffix pattern already pinned the index to 0
			if len(msg.Uid) != 36 || len(msg.Message) < 8 {
				continue
			}
			if !slp.HasSlpLokad(msg.Message[8:]) {
				continue
			}
			var txHash [32]byte
			copy(txHash[:], jutil.ByteReverse(msg.Uid[:32]))
			candidates = append(candidates, txHash)
		}
		if b.Verbose {
			log.Printf("slp validity backfill scan shard %d: %d candidates\n", shard, len(candidates))
		}
		if len(messages) < slpBackfillScanLimit {
			break // short page: shard range exhausted
		}
		start = jutil.CombineBytes(messages[len(messages)-1].Uid, []byte{0x00})
	}
	var undecided [][32]byte
	for i := 0; i < len(candidates); i += slpBackfillBatchSize {
		end := i + slpBackfillBatchSize
		if end > len(candidates) {
			end = len(candidates)
		}
		batch := candidates[i:end]
		validities, err := item_slp.GetValidities(b.Ctx, batch)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("error getting validities for slp candidates; %w", err)
		}
		var decided = make(map[[32]byte]bool, len(validities))
		for _, validity := range validities {
			decided[validity.TxHash] = true
		}
		for _, txHash := range batch {
			if !decided[txHash] {
				undecided = append(undecided, txHash)
			}
		}
	}
	return undecided, int64(len(candidates)), int64(len(candidates) - len(undecided)), nil
}

type orderedTx struct {
	txHash    [32]byte
	blockHash [32]byte
	height    int64
	index     uint32 // in-block index: stable tiebreak only, the topo sort provides correctness
}

// order maps each candidate to (block height, in-block index) and sorts.
// A tx in multiple blocks (orphans) uses its lowest block height that has a
// height row, tie-broken by block hash for determinism; txs with no mined
// block land in the mempool tail.
//
// Known edge: if a child's minimum height comes from an orphaned block with a
// height row while its parent's corresponding fork block lacks one, the child
// can sort into an earlier chunk than its parent, resolve ParentUnknown, and
// stay pending for this run (mined chunks don't cascade). Harmless: a re-run
// recollects still-undecided txs and the parent is decided by then.
func (b *SlpValidityBackfill) order(txHashes [][32]byte) ([]orderedTx, [][32]byte, error) {
	type blockRef struct {
		blockHash [32]byte
		index     uint32
	}
	var refs = make(map[[32]byte][]blockRef, len(txHashes))
	var blockSet = make(map[[32]byte]bool)
	for i := 0; i < len(txHashes); i += slpBackfillBatchSize {
		end := i + slpBackfillBatchSize
		if end > len(txHashes) {
			end = len(txHashes)
		}
		txBlocks, err := chain.GetTxBlocks(b.Ctx, txHashes[i:end])
		if err != nil {
			return nil, nil, fmt.Errorf("error getting tx blocks for slp backfill; %w", err)
		}
		for _, txBlock := range txBlocks {
			refs[txBlock.TxHash] = append(refs[txBlock.TxHash], blockRef{
				blockHash: txBlock.BlockHash,
				index:     txBlock.Index,
			})
			blockSet[txBlock.BlockHash] = true
		}
	}
	var blockHashes = make([][32]byte, 0, len(blockSet))
	for blockHash := range blockSet {
		blockHashes = append(blockHashes, blockHash)
	}
	var heights = make(map[[32]byte]int64, len(blockHashes))
	for i := 0; i < len(blockHashes); i += slpBackfillBatchSize {
		end := i + slpBackfillBatchSize
		if end > len(blockHashes) {
			end = len(blockHashes)
		}
		blockHeights, err := chain.GetBlockHeights(b.Ctx, blockHashes[i:end])
		if err != nil {
			return nil, nil, fmt.Errorf("error getting block heights for slp backfill; %w", err)
		}
		for _, blockHeight := range blockHeights {
			if height, ok := heights[blockHeight.BlockHash]; !ok || blockHeight.Height < height {
				heights[blockHeight.BlockHash] = blockHeight.Height
			}
		}
	}
	var ordered = make([]orderedTx, 0, len(txHashes))
	var mempool [][32]byte
	for _, txHash := range txHashes {
		var best *orderedTx
		for _, ref := range refs[txHash] {
			height, ok := heights[ref.blockHash]
			if !ok {
				continue
			}
			if best == nil || height < best.height ||
				(height == best.height && bytes.Compare(ref.blockHash[:], best.blockHash[:]) < 0) {
				best = &orderedTx{txHash: txHash, blockHash: ref.blockHash, height: height, index: ref.index}
			}
		}
		if best == nil {
			mempool = append(mempool, txHash)
			continue
		}
		ordered = append(ordered, *best)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].height != ordered[j].height {
			return ordered[i].height < ordered[j].height
		}
		if blockCmp := bytes.Compare(ordered[i].blockHash[:], ordered[j].blockHash[:]); blockCmp != 0 {
			return blockCmp < 0
		}
		if ordered[i].index != ordered[j].index {
			return ordered[i].index < ordered[j].index
		}
		return bytes.Compare(ordered[i].txHash[:], ordered[j].txHash[:]) < 0
	})
	return ordered, mempool, nil
}

// chunkByBlock splits the height-ordered txs into ~chunk-target pieces without
// ever splitting a block: CTOR means a same-block parent could otherwise land
// in a later chunk and stay pending forever. A giant block is one giant chunk.
func chunkByBlock(ordered []orderedTx) [][]orderedTx {
	var chunks [][]orderedTx
	var chunk []orderedTx
	for i, tx := range ordered {
		chunk = append(chunk, tx)
		blockEnd := i+1 == len(ordered) || ordered[i+1].blockHash != tx.blockHash
		if blockEnd && len(chunk) >= slpBackfillChunkTarget {
			chunks = append(chunks, chunk)
			chunk = nil
		}
	}
	if len(chunk) > 0 {
		chunks = append(chunks, chunk)
	}
	return chunks
}

// processChunk reconstructs, transcribes (one batched save), topologically
// sorts, and validates a chunk in a single call: with parents first and rows
// already written, in-call verdict visibility resolves arbitrarily deep
// same-chunk chains without cascade rounds.
func (b *SlpValidityBackfill) processChunk(validator *slp_validate.Validator, chunk []orderedTx) error {
	var txHashes = make([][32]byte, len(chunk))
	for i, tx := range chunk {
		txHashes[i] = tx.txHash
	}
	slpTxs, missing, err := slp_validate.ReconstructSlpTxs(b.Ctx, txHashes)
	if err != nil {
		return fmt.Errorf("error reconstructing txs for slp backfill; %w", err)
	}
	b.Missing += int64(len(missing))
	if b.Verbose {
		for _, missingHash := range missing {
			log.Printf("slp validity backfill skipping incomplete tx: %s\n", chainhash.Hash(missingHash))
		}
	}
	if len(slpTxs) == 0 {
		return nil
	}
	slpTxs = slp_validate.SortTopological(slpTxs)
	if err := slp_validate.TranscribeTxs(slpTxs); err != nil {
		return err
	}
	b.SlpTxs += int64(len(slpTxs))
	result, err := validator.ValidateTxs(b.Ctx, slpTxs)
	if err != nil {
		return fmt.Errorf("error validating txs for slp backfill; %w", err)
	}
	b.Valid += int64(result.Valid)
	b.Invalid += int64(result.Invalid)
	b.Pending += int64(result.Pending)
	return nil
}

// tail validates the unmined candidates with the cascade, which handles
// inter-mempool spend chains regardless of visit order.
func (b *SlpValidityBackfill) tail(txHashes [][32]byte) error {
	for i := 0; i < len(txHashes); i += slpSweepBatchSize {
		end := i + slpSweepBatchSize
		if end > len(txHashes) {
			end = len(txHashes)
		}
		slpTxs, missing, err := slp_validate.ReconstructSlpTxs(b.Ctx, txHashes[i:end])
		if err != nil {
			return fmt.Errorf("error reconstructing mempool txs for slp backfill; %w", err)
		}
		b.Missing += int64(len(missing))
		if len(slpTxs) == 0 {
			continue
		}
		if err := slp_validate.TranscribeTxs(slpTxs); err != nil {
			return err
		}
		b.SlpTxs += int64(len(slpTxs))
		for {
			result, err := slp_validate.ValidateTxsCascade(b.Ctx, slpTxs)
			if err != nil {
				return fmt.Errorf("error validating mempool txs for slp backfill; %w", err)
			}
			b.Valid += int64(result.Valid)
			b.Invalid += int64(result.Invalid)
			if result.Decided() == 0 {
				b.Pending += int64(result.Pending)
				break
			}
		}
	}
	return nil
}
