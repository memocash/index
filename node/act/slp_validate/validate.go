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

// add accumulates counts only: NewVerdicts is per-call cascade seed data, and
// accumulating it across a multi-million-verdict cascade is pure memory growth
// (no caller reads the aggregate).
func (r *Result) add(other *Result) {
	r.Valid += other.Valid
	r.Invalid += other.Invalid
	r.Pending += other.Pending
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
			// Defensive: callers only pass txs whose vout 0 is SLP (a tx with
			// lokad bytes only in later outputs is ignored upstream, never
			// reconstructed or handled here), so this should not be reached
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
	// Anything still unresolved is settled by the tx's vout-0 chain output:
	// validity is a property of the vout-0 message, so a known vout-0 script
	// without the SLP lokad definitively rules the tx out as SLP, while a
	// lokad-carrying (or unknown) one stays pending until its transcription
	// rows land. Chain rows are used rather than TxProcessed so conclusions
	// stay sound when the slp topics are rebuilt: wiped rows read as pending,
	// never as affirmatively not SLP.
	var voutZeroCheckSet = make(map[[32]byte]bool)
	for op := range outPointSet {
		if outputRows[op] == nil && batonRows[op] == nil && parentValidities[op.hash] == nil {
			voutZeroCheckSet[op.hash] = true
		}
	}
	for hash := range tokenSet {
		if genesisRows[hash] == nil {
			voutZeroCheckSet[hash] = true
		}
	}
	var notSlp = make(map[[32]byte]bool)
	if len(voutZeroCheckSet) > 0 {
		var voutZeroOuts = make([]memo.Out, 0, len(voutZeroCheckSet))
		for checkHash := range voutZeroCheckSet {
			hash := checkHash
			voutZeroOuts = append(voutZeroOuts, memo.Out{TxHash: hash[:], Index: 0})
		}
		txOutputs, err := chain.GetTxOutputs(ctx, voutZeroOuts)
		if err != nil {
			return fmt.Errorf("error getting vout-0 outputs for slp validate; %w", err)
		}
		for _, txOutput := range txOutputs {
			if !slp.HasSlpLokad(txOutput.LockScript) {
				notSlp[txOutput.TxHash] = true
			}
		}
	}
	var resolveTokenType = func(tokenHash [32]byte) (uint16, bool) {
		if genesis, ok := genesisRows[tokenHash]; ok {
			return uint16(genesis.TokenType), true
		}
		if notSlp[tokenHash] {
			// Not an SLP tx at all: the referenced token can never exist, so
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
			case validity != nil || notSlp[op.hash]:
				// A validity row implies the parent's transcription is
				// complete, so a missing row here is definitive; no SLP
				// lokad at vout 0 means the rows can never exist
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
