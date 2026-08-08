package maint

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/node/act/slp_validate"
	"github.com/memocash/index/node/obj/op_return"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
	"log"
)

const (
	slpSweepRawBatchSize = 1000
	slpSweepLogInterval  = 100
	// Heights per fetch chunk. Completeness comes from the exhaustive
	// per-shard pagination in chain.GetHeightBlocksRange, so this is purely
	// a batching/memory knob.
	slpSweepChunkHeights = 2000
)

// SlpValiditySweep walks blocks by height and validates every SLP tx to a
// fixpoint within each block (handles CTOR child-before-parent ordering). It
// is both the historical backfill and the safety net for anything the live
// save path leaves pending. Verdicts are final, so re-running is idempotent
// and already-decided txs are skipped cheaply.
type SlpValiditySweep struct {
	Ctx         context.Context
	Verbose     bool
	StartHeight int64 // 0 = resume from the saved cursor
	EndHeight   int64 // 0 = run to the tip

	Blocks  int64
	SlpTxs  int64
	Valid   int64
	Invalid int64
	Pending int64
	Missing int64
}

func NewSlpValiditySweep(ctx context.Context, verbose bool, startHeight, endHeight int64) *SlpValiditySweep {
	return &SlpValiditySweep{
		Ctx:         ctx,
		Verbose:     verbose,
		StartHeight: startHeight,
		EndHeight:   endHeight,
	}
}

func (s *SlpValiditySweep) getStartHeight() (int64, error) {
	if s.StartHeight > 0 {
		return s.StartHeight, nil
	}
	status, err := item.GetProcessStatus(s.Ctx, 0, item.ProcessStatusSlpValiditySweep)
	if err != nil {
		if client.IsMessageNotSetError(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting slp validity sweep status; %w", err)
	}
	if len(status.Status) >= 8 {
		return jutil.GetInt64Big(status.Status) + 1, nil
	}
	return 0, nil
}

func (s *SlpValiditySweep) Run() error {
	height, err := s.getStartHeight()
	if err != nil {
		return err
	}
	// The durable resume cursor only advances for runs that start from it;
	// an explicit --start is an ad hoc range and must not move it (a bounded
	// trial run would otherwise make a later default run skip everything
	// below its range).
	var persistCursor = s.StartHeight == 0
	if !persistCursor {
		log.Printf("explicit --start given: the saved resume cursor will not be updated\n")
	}
	// Sweep to the tip as of startup (or --end if lower). Later blocks are
	// validated by the live save path; re-running the sweep covers the rest.
	recentHeightBlock, err := chain.GetRecentHeightBlock(s.Ctx)
	if err != nil {
		return fmt.Errorf("error getting recent height block for slp validity sweep; %w", err)
	}
	if recentHeightBlock == nil {
		log.Printf("no height blocks found, nothing to sweep\n")
		return nil
	}
	var lastHeight = recentHeightBlock.Height
	if s.EndHeight > 0 && s.EndHeight < lastHeight {
		lastHeight = s.EndHeight
	}
	log.Printf("starting slp validity sweep at height %d (through %d)\n", height, lastHeight)
	var cursor = newSweepCursor()
	var intervalValid, intervalInvalid, intervalPending int64
	for chunkStart := height; chunkStart <= lastHeight; chunkStart += slpSweepChunkHeights {
		var chunkEnd = chunkStart + slpSweepChunkHeights
		if chunkEnd > lastHeight+1 {
			chunkEnd = lastHeight + 1
		}
		// Exhaustive per-shard range read: every record in [chunkStart,
		// chunkEnd) is guaranteed present, so heights in the chunk are safe
		// to checkpoint once passed
		heightBlocks, err := chain.GetHeightBlocksRange(s.Ctx, chunkStart, chunkEnd)
		if err != nil {
			return fmt.Errorf("error getting height blocks for slp validity sweep; %w", err)
		}
		for _, heightBlock := range heightBlocks {
			if completed, ok := cursor.advance(heightBlock.Height); ok {
				if err := s.saveCursor(persistCursor, completed); err != nil {
					return err
				}
			}
			if cursor.seen(heightBlock.BlockHash) {
				continue
			}
			if err := s.processBlock(heightBlock); err != nil {
				return fmt.Errorf("error processing block %s height %d for slp validity sweep; %w",
					chainhash.Hash(heightBlock.BlockHash), heightBlock.Height, err)
			}
			s.Blocks++
			if s.Blocks%slpSweepLogInterval == 0 {
				log.Printf("slp validity sweep height %d: interval valid: %d, invalid: %d, pending: %d\n",
					heightBlock.Height, s.Valid-intervalValid, s.Invalid-intervalInvalid, s.Pending-intervalPending)
				intervalValid, intervalInvalid, intervalPending = s.Valid, s.Invalid, s.Pending
			}
		}
	}
	if completed, ok := cursor.finish(); ok {
		if err := s.saveCursor(persistCursor, completed); err != nil {
			return err
		}
	}
	s.logTotals()
	return nil
}

// saveCursor persists a fully-completed height as the resume cursor. Resume
// starts at cursor+1, so a height is only saved once every block record at
// that height has been processed.
func (s *SlpValiditySweep) saveCursor(persist bool, height int64) error {
	if !persist {
		return nil
	}
	var status = item.NewProcessStatus(0, item.ProcessStatusSlpValiditySweep)
	status.Status = jutil.GetInt64DataBig(height)
	if err := status.Save(); err != nil {
		return fmt.Errorf("error saving slp validity sweep status; %w", err)
	}
	return nil
}

func (s *SlpValiditySweep) logTotals() {
	log.Printf("slp validity sweep done. blocks: %d, slp txs: %d, valid: %d, invalid: %d, pending: %d, missing: %d\n",
		s.Blocks, s.SlpTxs, s.Valid, s.Invalid, s.Pending, s.Missing)
}

func (s *SlpValiditySweep) processBlock(heightBlock *chain.HeightBlock) error {
	var txHashes [][32]byte
	var nextIndex uint32
	for {
		blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{
			BlockHash:  heightBlock.BlockHash,
			StartIndex: nextIndex,
			Limit:      client.LargeLimit,
		})
		if err != nil {
			return fmt.Errorf("error getting block txs for slp validity sweep; %w", err)
		}
		var maxIndex uint32
		var added int
		for _, blockTx := range blockTxs {
			if len(txHashes) > 0 && blockTx.Index < nextIndex {
				continue
			}
			txHashes = append(txHashes, blockTx.TxHash)
			added++
			if blockTx.Index > maxIndex {
				maxIndex = blockTx.Index
			}
		}
		if len(blockTxs) < client.LargeLimit {
			break
		}
		nextIndex = maxIndex + 1
		if added == 0 {
			break
		}
	}
	var slpTxs []slp_validate.Tx
	for i := 0; i < len(txHashes); i += slpSweepRawBatchSize {
		end := i + slpSweepRawBatchSize
		if end > len(txHashes) {
			end = len(txHashes)
		}
		batchSlpTxs, err := s.getSlpTxs(txHashes[i:end])
		if err != nil {
			return fmt.Errorf("error getting slp txs for slp validity sweep; %w", err)
		}
		slpTxs = append(slpTxs, batchSlpTxs...)
	}
	if len(slpTxs) == 0 {
		return nil
	}
	s.SlpTxs += int64(len(slpTxs))
	if err := s.transcribeUndecided(slpTxs); err != nil {
		return err
	}
	// Iterate to a fixpoint so in-block CTOR parent/child chains resolve no
	// matter the order they appear in the block
	for {
		result, err := slp_validate.ValidateTxs(s.Ctx, slpTxs)
		if err != nil {
			return fmt.Errorf("error validating txs for slp validity sweep; %w", err)
		}
		s.Valid += int64(result.Valid)
		s.Invalid += int64(result.Invalid)
		if result.Decided() == 0 {
			s.Pending += int64(result.Pending)
			return nil
		}
	}
}

// getSlpTxs reconstructs the given txs from chain topics and returns the ones
// carrying an SLP-lokad output. Incomplete txs are counted and skipped.
func (s *SlpValiditySweep) getSlpTxs(txHashes [][32]byte) ([]slp_validate.Tx, error) {
	txs, err := chain.GetTxsByHashes(s.Ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting txs; %w", err)
	}
	inputs, err := chain.GetTxInputsByHashes(s.Ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting tx inputs; %w", err)
	}
	outputs, err := chain.GetTxOutputsByHashes(s.Ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting tx outputs; %w", err)
	}
	var txMap = make(map[[32]byte]*chain.Tx, len(txs))
	for _, tx := range txs {
		txMap[tx.TxHash] = tx
	}
	var inMap = make(map[[32]byte][]*chain.TxInput)
	for _, input := range inputs {
		inMap[input.TxHash] = append(inMap[input.TxHash], input)
	}
	var outMap = make(map[[32]byte][]*chain.TxOutput)
	for _, output := range outputs {
		outMap[output.TxHash] = append(outMap[output.TxHash], output)
	}
	var slpTxs []slp_validate.Tx
	for _, txHash := range txHashes {
		tx, ok := txMap[txHash]
		if !ok || len(outMap[txHash]) == 0 {
			s.Missing++
			continue
		}
		var anySlp bool
		for _, output := range outMap[txHash] {
			if slp.HasSlpLokad(output.LockScript) {
				anySlp = true
				break
			}
		}
		if !anySlp {
			continue
		}
		msgTx := buildMsgTx(tx, inMap[txHash], outMap[txHash])
		if msgTx.TxHash() != chainhash.Hash(txHash) {
			// Rebuilt hash mismatch means the index data for this tx is incomplete
			s.Missing++
			if s.Verbose {
				log.Printf("slp validity sweep skipping incomplete tx: %s\n", chainhash.Hash(txHash))
			}
			continue
		}
		slpTxs = append(slpTxs, slp_validate.Tx{
			TxHash:  txHash,
			Inputs:  msgTx.TxIn,
			Outputs: msgTx.TxOut,
		})
	}
	return slpTxs, nil
}

// transcribeUndecided re-runs lenient SLP transcription for txs with no
// verdict yet, covering txs the live path missed (e.g. pre-fix txs skipped
// for lacking an input address, or non-minimal lokad pushes).
func (s *SlpValiditySweep) transcribeUndecided(slpTxs []slp_validate.Tx) error {
	var txHashes = make([][32]byte, 0, len(slpTxs))
	for _, tx := range slpTxs {
		txHashes = append(txHashes, tx.TxHash)
	}
	validities, err := item_slp.GetValidities(s.Ctx, txHashes)
	if err != nil {
		return fmt.Errorf("error getting validities for slp validity sweep transcription; %w", err)
	}
	var decided = make(map[[32]byte]bool, len(validities))
	for _, validity := range validities {
		decided[validity.TxHash] = true
	}
	for _, tx := range slpTxs {
		if decided[tx.TxHash] {
			continue
		}
		for index, output := range tx.Outputs {
			if !slp.HasSlpLokad(output.PkScript) {
				continue
			}
			pushData, err := txscript.PushedData(output.PkScript)
			if err != nil {
				continue
			}
			if err := op_return.TranscribeSlp(parse.OpReturn{
				Saved:       true,
				TxHash:      tx.TxHash,
				PushData:    pushData,
				Outputs:     tx.Outputs,
				Inputs:      tx.Inputs,
				OutputIndex: index,
			}); err != nil {
				return fmt.Errorf("error transcribing slp tx for validity sweep; %w", err)
			}
		}
	}
	return nil
}
