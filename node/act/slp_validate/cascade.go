package slp_validate

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/item/chain"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
	"log"
	"sort"
	"time"
)

// ValidateTxsCascade validates txs and then walks forward through the spend
// graph: the spenders of each newly decided tx's outputs are looked up
// (chain_output_input) and any still-pending SLP spenders are validated in
// turn, repeating until no new verdicts land. A decided ancestor therefore
// immediately resolves pending descendants of any depth, without waiter
// machinery: spender rows are written by the first saver, before validation,
// so either a child's own validation sees the parent's verdict or the
// parent's spender lookup sees the child.
// cascadeLogInterval is how often a still-running cascade reports progress; a
// big token graph can keep one cascade busy for hours, and without a heartbeat
// that is indistinguishable from a hang.
const cascadeLogInterval = time.Minute

func ValidateTxsCascade(ctx context.Context, txs []Tx) (*Result, error) {
	var total = new(Result)
	var current = txs
	// Transcription depends only on the tx itself, so each spender is
	// transcribed at most once per cascade; a pending spender that reappears
	// in later rounds (another of its parents decided) skips straight to
	// re-validation instead of redundantly rewriting its rows every round
	var transcribed = make(map[[32]byte]bool)
	var lastLog = time.Now()
	for round := 1; len(current) > 0; round++ {
		result, err := ValidateTxs(ctx, current)
		if err != nil {
			return nil, err
		}
		total.add(result)
		if len(result.NewVerdicts) == 0 {
			break
		}
		current, err = getPendingSpenders(ctx, result.NewVerdicts, transcribed)
		if err != nil {
			return nil, err
		}
		if time.Since(lastLog) >= cascadeLogInterval {
			log.Printf("slp validity cascade at round %d: frontier %d, valid %d, invalid %d, pending %d\n",
				round, len(current), total.Valid, total.Invalid, total.Pending)
			lastLog = time.Now()
		}
	}
	return total, nil
}

// getPendingSpenders returns the reconstructed SLP txs spending outputs of
// the given txs that do not have a verdict yet, transcribing any not already
// transcribed this cascade. Spenders whose chain rows are incomplete
// (mid-save) are skipped; their own save-time validation covers them.
func getPendingSpenders(ctx context.Context, txHashes [][32]byte, transcribed map[[32]byte]bool) ([]Tx, error) {
	outputInputs, err := chain.GetOutputInputsForTxHashes(ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting output inputs for slp cascade; %w", err)
	}
	var spenderSet = make(map[[32]byte]bool)
	for _, outputInput := range outputInputs {
		spenderSet[outputInput.Hash] = true
	}
	if len(spenderSet) == 0 {
		return nil, nil
	}
	var spenderHashes = make([][32]byte, 0, len(spenderSet))
	for spenderHash := range spenderSet {
		spenderHashes = append(spenderHashes, spenderHash)
	}
	validities, err := item_slp.GetValidities(ctx, spenderHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting spender validities for slp cascade; %w", err)
	}
	for _, validity := range validities {
		delete(spenderSet, validity.TxHash)
	}
	var undecided = make([][32]byte, 0, len(spenderSet))
	for spenderHash := range spenderSet {
		undecided = append(undecided, spenderHash)
	}
	if len(undecided) == 0 {
		return nil, nil
	}
	slpTxs, _, err := ReconstructSlpTxs(ctx, undecided)
	if err != nil {
		return nil, fmt.Errorf("error reconstructing spender txs for slp cascade; %w", err)
	}
	// Spenders found by the cascade may not have been transcribed yet (the
	// audit sweep or live path may simply not have reached them); their
	// verdicts land now, so their rows must too, or descendants would read
	// the missing rows as contributing nothing and go falsely invalid
	var toTranscribe = make([]Tx, 0, len(slpTxs))
	for _, slpTx := range slpTxs {
		if !transcribed[slpTx.TxHash] {
			transcribed[slpTx.TxHash] = true
			toTranscribe = append(toTranscribe, slpTx)
		}
	}
	if err := CascadeTranscribe(toTranscribe); err != nil {
		return nil, fmt.Errorf("error transcribing spender txs for slp cascade; %w", err)
	}
	return slpTxs, nil
}

// ReconstructSlpTxs rebuilds the given txs from the chain topics and returns
// the ones carrying an SLP-lokad output, plus the hashes of txs that could
// not be fully reconstructed.
func ReconstructSlpTxs(ctx context.Context, txHashes [][32]byte) ([]Tx, [][32]byte, error) {
	txs, err := chain.GetTxsByHashes(ctx, txHashes)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting txs for slp reconstruct; %w", err)
	}
	inputs, err := chain.GetTxInputsByHashes(ctx, txHashes)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting tx inputs for slp reconstruct; %w", err)
	}
	outputs, err := chain.GetTxOutputsByHashes(ctx, txHashes)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting tx outputs for slp reconstruct; %w", err)
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
	var slpTxs []Tx
	var missing [][32]byte
	for _, txHash := range txHashes {
		tx, ok := txMap[txHash]
		if !ok || len(outMap[txHash]) == 0 {
			missing = append(missing, txHash)
			continue
		}
		// An SLP action is defined only by the vout-0 message; a tx whose
		// vout 0 carries no lokad is not SLP even if a later output does
		// (spec Consideration A), so it contributes nothing as a spender and
		// needs no verdict to drop out of the cascade
		var voutZero *chain.TxOutput
		for _, output := range outMap[txHash] {
			if output.Index == 0 {
				voutZero = output
				break
			}
		}
		if voutZero == nil || !slp.HasSlpLokad(voutZero.LockScript) {
			continue
		}
		msgTx := buildTxMsg(tx, inMap[txHash], outMap[txHash])
		if msgTx.TxHash() != chainhash.Hash(txHash) {
			// Rebuilt hash mismatch means the chain rows are incomplete
			missing = append(missing, txHash)
			continue
		}
		slpTxs = append(slpTxs, Tx{
			TxHash:  txHash,
			Inputs:  msgTx.TxIn,
			Outputs: msgTx.TxOut,
		})
	}
	return slpTxs, missing, nil
}

func buildTxMsg(tx *chain.Tx, inputs []*chain.TxInput, outputs []*chain.TxOutput) *wire.MsgTx {
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].Index < inputs[j].Index })
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].Index < outputs[j].Index })
	msgTx := wire.NewMsgTx(tx.Version)
	msgTx.LockTime = tx.LockTime
	for _, input := range inputs {
		prevHash := chainhash.Hash(input.PrevHash)
		outPoint := wire.NewOutPoint(&prevHash, input.PrevIndex)
		txIn := wire.NewTxIn(outPoint, input.UnlockScript)
		txIn.Sequence = input.Sequence
		msgTx.AddTxIn(txIn)
	}
	for _, output := range outputs {
		msgTx.AddTxOut(wire.NewTxOut(output.Value, output.LockScript))
	}
	return msgTx
}
