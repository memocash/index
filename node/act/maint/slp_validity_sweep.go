package maint

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/node/act/slp_validate"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
	"github.com/memocash/index/ref/config"
	"log"
)

const (
	slpSweepBatchSize       = 1000
	slpSweepLogPageInterval = 100
)

// SlpValiditySweep validates SLP txs without a verdict, driven by the index's
// own datasets rather than a chain walk.
//
// Default (index sweep): iterates the slp genesis/mint/send topics — every
// transcribed SLP tx — and validates the undecided ones. Near-free on a
// healthy index; this is the routine safety net for txs the live save path
// left pending.
//
// Audit: scans every chain tx output for an SLP-lokad lock script, leniently
// transcribes anything the live path missed, and validates the undecided.
// This is the historical backfill and deep audit. A per-shard uid cursor is
// checkpointed after each page so an interrupted run resumes where it left
// off; the cursor is cleared when the shard completes, since new txs land
// anywhere in uid order and a finished cursor must not carry into the next
// audit. The cursor lives in the process_status topic, NOT the slp topics, so
// it survives a wipe-and-repopulate: an audit meant to rebuild wiped slp data
// must run with Fresh set, or it silently skips everything before the dead
// run's cursor.
//
// Verdicts are final, so both modes are idempotent, and visit order does not
// matter: a child visited before its parent parks pending and is resolved by
// the cascade when the parent's verdict lands.
type SlpValiditySweep struct {
	Ctx     context.Context
	Verbose bool
	Audit   bool
	Fresh   bool // audit only: durably clear saved resume cursors up front, start from the beginning

	Checked int64 // rows scanned: slp topic rows, or tx outputs in audit mode
	SlpTxs  int64 // undecided slp txs validated
	Valid   int64
	Invalid int64
	Pending int64
	Missing int64
}

func NewSlpValiditySweep(ctx context.Context, verbose, audit bool) *SlpValiditySweep {
	return &SlpValiditySweep{
		Ctx:     ctx,
		Verbose: verbose,
		Audit:   audit,
	}
}

func (s *SlpValiditySweep) Run() error {
	if s.Audit && s.Fresh {
		// Durably clear before any page work: if the cursors were only
		// ignored in memory, a fresh run killed before its first checkpoint
		// would leave the stale cursor in place, and a plain retry would
		// resume from it and silently skip the wiped prefix again
		if err := s.ClearAuditCursors(); err != nil {
			return fmt.Errorf("error clearing audit cursors for fresh slp validity audit; %w", err)
		}
	}
	for _, shardConfig := range config.GetQueueShards() {
		if s.Audit {
			if err := s.auditShard(shardConfig.Shard); err != nil {
				return fmt.Errorf("error auditing tx outputs for slp validity shard %d; %w", shardConfig.Shard, err)
			}
		} else if err := s.sweepShard(shardConfig.Shard); err != nil {
			return fmt.Errorf("error sweeping slp topics for slp validity shard %d; %w", shardConfig.Shard, err)
		}
	}
	log.Printf("slp validity sweep done. checked: %d, slp txs: %d, valid: %d, invalid: %d, pending: %d, missing: %d\n",
		s.Checked, s.SlpTxs, s.Valid, s.Invalid, s.Pending, s.Missing)
	return nil
}

// sweepShard iterates the slp transcription topics for a shard and validates
// the txs among them with no verdict.
func (s *SlpValiditySweep) sweepShard(shard uint32) error {
	for _, topic := range []string{db.TopicSlpGenesis, db.TopicSlpMint, db.TopicSlpSend} {
		var startUid []byte
		for {
			txHashes, err := item_slp.GetTopicTxHashes(s.Ctx, topic, shard, startUid)
			if err != nil {
				return fmt.Errorf("error getting tx hashes for slp topic: %s; %w", topic, err)
			}
			s.Checked += int64(len(txHashes))
			if err := s.validateUndecided(txHashes); err != nil {
				return err
			}
			if len(txHashes) < client.HugeLimit {
				break
			}
			startUid = jutil.ByteReverse(txHashes[len(txHashes)-1][:])
		}
		if s.Verbose {
			log.Printf("slp validity sweep finished topic %s shard %d\n", topic, shard)
		}
	}
	return nil
}

// ClearAuditCursors durably removes every shard's saved audit resume cursor,
// so the next audit (this run or a retry after any interruption) starts from
// the beginning. Run calls it before any page work when Fresh is set.
func (s *SlpValiditySweep) ClearAuditCursors() error {
	for _, shardConfig := range config.GetQueueShards() {
		status, err := item.GetProcessStatus(s.Ctx, uint(shardConfig.Shard), item.ProcessStatusSlpValiditySweep)
		if err != nil {
			if client.IsMessageNotSetError(err) {
				continue
			}
			return fmt.Errorf("error getting slp validity audit status for shard %d; %w", shardConfig.Shard, err)
		}
		if len(status.Status) == 0 {
			continue
		}
		log.Printf("clearing saved slp validity cursor for shard %d (fresh audit): %x\n",
			shardConfig.Shard, status.Status)
		status.Status = nil
		if err := status.Save(); err != nil {
			return fmt.Errorf("error clearing slp validity cursor for shard %d; %w", shardConfig.Shard, err)
		}
	}
	return nil
}

// auditShard scans the shard's chain tx outputs for SLP lokad scripts and
// validates any carrying tx with no verdict, transcribing it first.
func (s *SlpValiditySweep) auditShard(shard uint32) error {
	var startUid []byte
	status, err := item.GetProcessStatus(s.Ctx, uint(shard), item.ProcessStatusSlpValiditySweep)
	if err != nil {
		if !client.IsMessageNotSetError(err) {
			return fmt.Errorf("error getting slp validity audit status; %w", err)
		}
		status = item.NewProcessStatus(uint(shard), item.ProcessStatusSlpValiditySweep)
	} else if len(status.Status) == memo.TxHashLength+4 {
		startUid = status.Status
		log.Printf("resuming slp validity audit shard %d from cursor %x\n", shard, startUid)
	} else if len(status.Status) > 0 {
		// Not a tx output uid: the pre-dataset sweeper stored an 8-byte block
		// height under this status name. Resuming from a wrong-format key
		// would silently skip outputs, so start from the beginning instead
		log.Printf("ignoring legacy slp validity cursor for shard %d (%x), starting audit from the beginning\n",
			shard, status.Status)
	}
	for page := 1; ; page++ {
		txOutputs, err := chain.GetAllTxOutputs(s.Ctx, shard, startUid)
		if err != nil {
			return fmt.Errorf("error getting tx outputs for slp validity audit; %w", err)
		}
		var txHashes [][32]byte
		var seen = make(map[[32]byte]struct{})
		for _, txOutput := range txOutputs {
			if uid := txOutput.GetUid(); jutil.ByteGT(uid, startUid) {
				startUid = uid
			}
			s.Checked++
			// An SLP action lives only at vout 0; lokad bytes in any later
			// output are not SLP (spec Consideration A), so only vout 0 makes
			// a tx an audit candidate
			if txOutput.Index != 0 || !slp.HasSlpLokad(txOutput.LockScript) {
				continue
			}
			if _, ok := seen[txOutput.TxHash]; ok {
				continue
			}
			seen[txOutput.TxHash] = struct{}{}
			txHashes = append(txHashes, txOutput.TxHash)
		}
		if err := s.validateUndecided(txHashes); err != nil {
			return err
		}
		var done = len(txOutputs) < client.HugeLimit
		if done {
			// Cleared cursor: the next audit starts from the beginning
			status.Status = nil
		} else {
			status.Status = startUid
		}
		if err := status.Save(); err != nil {
			return fmt.Errorf("error saving slp validity audit status; %w", err)
		}
		if done {
			return nil
		}
		if page%slpSweepLogPageInterval == 0 {
			log.Printf("slp validity audit shard %d at %x: checked %d, slp txs %d, valid %d, invalid %d\n",
				shard, startUid, s.Checked, s.SlpTxs, s.Valid, s.Invalid)
		}
	}
}

// validateUndecided validates the given txs that have no verdict yet,
// re-running lenient transcription for them first (covers txs the live path
// missed, e.g. pre-fix txs skipped for lacking an input address, or
// non-minimal lokad pushes). Cascading validation resolves chains regardless
// of visit order (and any pending descendants); the fixpoint loop remains as
// a backstop.
func (s *SlpValiditySweep) validateUndecided(txHashes [][32]byte) error {
	for i := 0; i < len(txHashes); i += slpSweepBatchSize {
		end := i + slpSweepBatchSize
		if end > len(txHashes) {
			end = len(txHashes)
		}
		batch := txHashes[i:end]
		validities, err := item_slp.GetValidities(s.Ctx, batch)
		if err != nil {
			return fmt.Errorf("error getting validities for slp validity sweep; %w", err)
		}
		var decided = make(map[[32]byte]bool, len(validities))
		for _, validity := range validities {
			decided[validity.TxHash] = true
		}
		var undecided = make([][32]byte, 0, len(batch))
		for _, txHash := range batch {
			if !decided[txHash] {
				undecided = append(undecided, txHash)
			}
		}
		if len(undecided) == 0 {
			continue
		}
		slpTxs, missing, err := slp_validate.ReconstructSlpTxs(s.Ctx, undecided)
		if err != nil {
			return fmt.Errorf("error reconstructing txs for slp validity sweep; %w", err)
		}
		s.Missing += int64(len(missing))
		if s.Verbose {
			for _, missingHash := range missing {
				log.Printf("slp validity sweep skipping incomplete tx: %s\n", chainhash.Hash(missingHash))
			}
		}
		if len(slpTxs) == 0 {
			continue
		}
		s.SlpTxs += int64(len(slpTxs))
		if err := slp_validate.TranscribeTxs(slpTxs); err != nil {
			return err
		}
		for {
			result, err := slp_validate.ValidateTxsCascade(s.Ctx, slpTxs)
			if err != nil {
				return fmt.Errorf("error validating txs for slp validity sweep; %w", err)
			}
			s.Valid += int64(result.Valid)
			s.Invalid += int64(result.Invalid)
			if result.Decided() == 0 {
				s.Pending += int64(result.Pending)
				break
			}
		}
	}
	return nil
}
