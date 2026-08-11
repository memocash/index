package slp_validate

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/item/chain"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
	"sort"
)

// ValidateTxsCascade validates txs and then walks forward through the spend
// graph: the spenders of each newly decided tx's outputs are looked up
// (chain_output_input) and any still-pending SLP spenders are validated in
// turn, repeating until no new verdicts land. A decided ancestor therefore
// immediately resolves pending descendants of any depth, without waiter
// machinery: spender rows are written by the first saver, before validation,
// so either a child's own validation sees the parent's verdict or the
// parent's spender lookup sees the child.
func ValidateTxsCascade(ctx context.Context, txs []Tx) (*Result, error) {
	var total = new(Result)
	var current = txs
	for len(current) > 0 {
		result, err := ValidateTxs(ctx, current)
		if err != nil {
			return nil, err
		}
		total.add(result)
		if len(result.NewVerdicts) == 0 {
			break
		}
		current, err = getPendingSpenders(ctx, result.NewVerdicts)
		if err != nil {
			return nil, err
		}
	}
	return total, nil
}

// getPendingSpenders returns the reconstructed SLP txs spending outputs of
// the given txs that do not have a verdict yet. Spenders whose chain rows are
// incomplete (mid-save) are skipped; their own save-time validation covers
// them.
func getPendingSpenders(ctx context.Context, txHashes [][32]byte) ([]Tx, error) {
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
