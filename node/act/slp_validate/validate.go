// Package slp_validate resolves the inputs of SLP transactions against the
// index and applies the pure validator (ref/bitcoin/tx/slp), writing
// slp.Validity verdicts. Verdicts are only written on affirmative evidence;
// anything unresolved stays pending (no row) and is resolved later by the
// cascade when an ancestor's verdict lands, or by the slp-validity-sweep
// maintenance command.
package slp_validate

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
)

type Tx struct {
	TxHash  [32]byte
	Inputs  []*wire.TxIn
	Outputs []*wire.TxOut
}

type Result struct {
	Valid   int
	Invalid int
	Pending int
	// NewVerdicts holds the tx hashes decided by this call, the seeds for
	// cascading validation to their pending spenders
	NewVerdicts [][32]byte
}

// Decided returns the number of new verdicts written this call, used by the
// sweeper's fixpoint loop.
func (r *Result) Decided() int {
	return r.Valid + r.Invalid
}

func (r *Result) add(other *Result) {
	r.Valid += other.Valid
	r.Invalid += other.Invalid
	r.Pending += other.Pending
	r.NewVerdicts = append(r.NewVerdicts, other.NewVerdicts...)
}

type outPoint struct {
	hash  [32]byte
	index uint32
}

type work struct {
	tx  Tx
	msg *slp.Msg
}

// ValidateTxs validates a batch of SLP-carrying txs, resolving all inputs with
// batched reads, and saves verdicts for txs that are decidable. Txs that
// already have a verdict are skipped (verdicts are final).
func ValidateTxs(ctx context.Context, txs []Tx) (*Result, error) {
	var result = new(Result)
	var txHashes = make([][32]byte, 0, len(txs))
	for _, tx := range txs {
		txHashes = append(txHashes, tx.TxHash)
	}
	existingValidities, err := item_slp.GetValidities(ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting existing slp validities; %w", err)
	}
	var existing = make(map[[32]byte]bool, len(existingValidities))
	for _, validity := range existingValidities {
		existing[validity.TxHash] = true
	}
	var verdicts = make(map[[32]byte]slp.Verdict)
	var works []*work
	for _, tx := range txs {
		if existing[tx.TxHash] {
			continue
		}
		if len(tx.Outputs) == 0 || !slp.HasSlpLokad(tx.Outputs[0].PkScript) {
			// The SLP message that triggered handling is not at vout 0
			verdicts[tx.TxHash] = slp.Verdict{Status: slp.StatusInvalid, Reason: slp.ReasonNotVoutZero}
			continue
		}
		msg, err := slp.Parse(tx.Outputs[0].PkScript)
		if err != nil {
			verdicts[tx.TxHash] = slp.Verdict{Status: slp.StatusInvalid, Reason: slp.ReasonParse}
			continue
		}
		if msg.TxType == memo.SlpTxTypeCommit {
			continue
		}
		works = append(works, &work{tx: tx, msg: msg})
	}
	if err := resolveAndValidate(ctx, works, verdicts); err != nil {
		return nil, err
	}
	var objects []db.Object
	for txHash, verdict := range verdicts {
		switch verdict.Status {
		case slp.StatusValid:
			result.Valid++
		case slp.StatusInvalid:
			result.Invalid++
		default:
			result.Pending++
			continue
		}
		objects = append(objects, &item_slp.Validity{
			TxHash: txHash,
			Status: uint8(verdict.Status),
			Reason: uint8(verdict.Reason),
		})
		result.NewVerdicts = append(result.NewVerdicts, txHash)
	}
	if len(objects) > 0 {
		if err := db.Save(objects); err != nil {
			return nil, fmt.Errorf("error saving slp validities; %w", err)
		}
	}
	return result, nil
}

func resolveAndValidate(ctx context.Context, works []*work, verdicts map[[32]byte]slp.Verdict) error {
	if len(works) == 0 {
		return nil
	}
	// Batch-read SLP output and baton rows for every prevout
	var outPointSet = make(map[outPoint]bool)
	var parentSet = make(map[[32]byte]bool)
	for _, w := range works {
		for _, txIn := range w.tx.Inputs {
			var op = outPoint{hash: txIn.PreviousOutPoint.Hash, index: txIn.PreviousOutPoint.Index}
			outPointSet[op] = true
			parentSet[op.hash] = true
		}
	}
	var outs = make([]memo.Out, 0, len(outPointSet))
	for op := range outPointSet {
		hash := op.hash
		outs = append(outs, memo.Out{TxHash: hash[:], Index: op.index})
	}
	var outputRows = make(map[outPoint]*item_slp.Output)
	var batonRows = make(map[outPoint]*item_slp.Baton)
	if len(outs) > 0 {
		outputs, err := item_slp.GetOutputs(ctx, outs)
		if err != nil {
			return fmt.Errorf("error getting slp outputs for validate; %w", err)
		}
		for _, output := range outputs {
			outputRows[outPoint{hash: output.TxHash, index: output.Index}] = output
		}
		batons, err := item_slp.GetBatons(ctx, outs)
		if err != nil {
			return fmt.Errorf("error getting slp batons for validate; %w", err)
		}
		for _, baton := range batons {
			batonRows[outPoint{hash: baton.TxHash, index: baton.Index}] = baton
		}
	}
	// Batch-read parent verdicts
	var parentHashes = make([][32]byte, 0, len(parentSet))
	for hash := range parentSet {
		parentHashes = append(parentHashes, hash)
	}
	var parentValidities = make(map[[32]byte]*item_slp.Validity)
	if len(parentHashes) > 0 {
		validities, err := item_slp.GetValidities(ctx, parentHashes)
		if err != nil {
			return fmt.Errorf("error getting parent slp validities; %w", err)
		}
		for _, validity := range validities {
			parentValidities[validity.TxHash] = validity
		}
	}
	// Each spent SLP row's token type comes from that token's genesis: for
	// valid parent chains the genesis-declared type and the parent-declared
	// type are always equal, and only valid parents contribute
	var tokenSet = make(map[[32]byte]bool)
	for _, output := range outputRows {
		tokenSet[output.TokenHash] = true
	}
	for _, baton := range batonRows {
		tokenSet[baton.TokenHash] = true
	}
	var genesisRows = make(map[[32]byte]*item_slp.Genesis)
	if len(tokenSet) > 0 {
		var tokenHashes = make([][32]byte, 0, len(tokenSet))
		for hash := range tokenSet {
			tokenHashes = append(tokenHashes, hash)
		}
		geneses, err := item_slp.GetGeneses(ctx, tokenHashes)
		if err != nil {
			return fmt.Errorf("error getting slp geneses for validate; %w", err)
		}
		for _, genesis := range geneses {
			genesisRows[genesis.TxHash] = genesis
		}
	}
	// Anything still unresolved needs TxProcessed to distinguish
	// "genuinely not SLP" from "not ingested yet"
	var processedCheckSet = make(map[[32]byte]bool)
	for op := range outPointSet {
		if outputRows[op] == nil && batonRows[op] == nil && parentValidities[op.hash] == nil {
			processedCheckSet[op.hash] = true
		}
	}
	for hash := range tokenSet {
		if genesisRows[hash] == nil {
			processedCheckSet[hash] = true
		}
	}
	var processed = make(map[[32]byte]bool)
	if len(processedCheckSet) > 0 {
		var processedHashes = make([][32]byte, 0, len(processedCheckSet))
		for hash := range processedCheckSet {
			processedHashes = append(processedHashes, hash)
		}
		txProcesseds, err := chain.GetTxProcessed(ctx, processedHashes)
		if err != nil {
			return fmt.Errorf("error getting tx processed for slp validate; %w", err)
		}
		for _, txProcessed := range txProcesseds {
			var hash [32]byte
			copy(hash[:], txProcessed.TxHash)
			processed[hash] = true
		}
	}
	var resolveTokenType = func(tokenHash [32]byte) (uint16, bool) {
		if genesis, ok := genesisRows[tokenHash]; ok {
			return uint16(genesis.TokenType), true
		}
		if processed[tokenHash] {
			// Processed with no genesis row: affirmatively not a token, so
			// the row can never match a handled type and contributes zero
			return 0, true
		}
		return 0, false
	}
	for _, w := range works {
		var inputs = make([]slp.Input, len(w.tx.Inputs))
		for i, txIn := range w.tx.Inputs {
			var op = outPoint{hash: txIn.PreviousOutPoint.Hash, index: txIn.PreviousOutPoint.Index}
			var input = slp.Input{PrevHash: op.hash, Index: op.index}
			outputRow, batonRow := outputRows[op], batonRows[op]
			if outputRow != nil {
				tokenType, typeKnown := resolveTokenType(outputRow.TokenHash)
				input.Output = &slp.InputToken{
					TokenHash: outputRow.TokenHash,
					TokenType: tokenType,
					TypeKnown: typeKnown,
					Quantity:  outputRow.Quantity,
				}
			}
			if batonRow != nil {
				tokenType, typeKnown := resolveTokenType(batonRow.TokenHash)
				input.Baton = &slp.InputBaton{
					TokenHash: batonRow.TokenHash,
					TokenType: tokenType,
					TypeKnown: typeKnown,
				}
			}
			validity := parentValidities[op.hash]
			switch {
			case outputRow != nil || batonRow != nil:
				switch {
				case validity == nil:
					input.State = slp.ParentPending
				case validity.IsValid():
					input.State = slp.ParentValid
				default:
					input.State = slp.ParentInvalid
				}
			case validity != nil || processed[op.hash]:
				// A validity row implies the parent's transcription is
				// complete, so a missing row here is definitive
				input.State = slp.ParentNotSlp
			default:
				input.State = slp.ParentUnknown
			}
			inputs[i] = input
		}
		verdicts[w.tx.TxHash] = slp.Validate(slp.TxData{
			Msg:    w.msg,
			Inputs: inputs,
		})
	}
	return nil
}
