package maint

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/node/act/slp_validate"
	"github.com/memocash/index/ref/config"
	"log"
)

const slpSweepBatchSize = 1000

// SlpValiditySweep validates SLP txs without a verdict, driven by the index's
// own datasets: it iterates the slp genesis/mint/send topics — every
// transcribed SLP tx — and validates the undecided ones. Near-free on a
// healthy index; this is the routine safety net for txs the live save path
// left pending. Historical backfill and deep audit (finding txs the live path
// never transcribed) is SlpValidityBackfill's job, which scans the raw chain
// tx outputs instead.
//
// Verdicts are final, so the sweep is idempotent, and visit order does not
// matter: a child visited before its parent parks pending and is resolved by
// the cascade when the parent's verdict lands.
type SlpValiditySweep struct {
	Ctx     context.Context
	Verbose bool

	Checked int64 // slp topic rows scanned
	SlpTxs  int64 // undecided slp txs validated
	Valid   int64
	Invalid int64
	Pending int64
	Missing int64
}

func NewSlpValiditySweep(ctx context.Context, verbose bool) *SlpValiditySweep {
	return &SlpValiditySweep{
		Ctx:     ctx,
		Verbose: verbose,
	}
}

func (s *SlpValiditySweep) Run() error {
	for _, shardConfig := range config.GetQueueShards() {
		if err := s.sweepShard(shardConfig.Shard); err != nil {
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
