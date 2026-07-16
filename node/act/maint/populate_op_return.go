package maint

import (
	"context"
	"fmt"
	"sort"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/dbi"
)

const populateOpReturnBatchSize = 1000

// PopulateOpReturn runs the op_return saver over txs that are already stored in the
// Index chain topics, reconstructing each tx from its stored version, inputs, and
// outputs. This populates any op_return-derived data (e.g. new memo item types) for
// known tx hashes without a BCH node scan.
type PopulateOpReturn struct {
	Ctx     context.Context
	Verbose bool
}

// PopulateOpReturnResult reports the outcome of a run. Missing holds tx hashes (in
// internal byte order) that were not fully present in the Index and so were skipped.
type PopulateOpReturnResult struct {
	Processed int
	Missing   [][32]byte
}

func NewPopulateOpReturn(ctx context.Context, verbose bool) *PopulateOpReturn {
	return &PopulateOpReturn{
		Ctx:     ctx,
		Verbose: verbose,
	}
}

func (p *PopulateOpReturn) Run(hashes [][32]byte) (*PopulateOpReturnResult, error) {
	var result = new(PopulateOpReturnResult)
	opReturnSaver := saver.NewOpReturn(p.Verbose)
	for i := 0; i < len(hashes); i += populateOpReturnBatchSize {
		end := i + populateOpReturnBatchSize
		if end > len(hashes) {
			end = len(hashes)
		}
		if err := p.processBatch(opReturnSaver, hashes[i:end], result); err != nil {
			return nil, fmt.Errorf("error processing batch for populate op return; %w", err)
		}
	}
	return result, nil
}

func (p *PopulateOpReturn) processBatch(opReturnSaver *saver.OpReturn, hashes [][32]byte, result *PopulateOpReturnResult) error {
	txs, err := chain.GetTxsByHashes(p.Ctx, hashes)
	if err != nil {
		return fmt.Errorf("error getting txs; %w", err)
	}
	seens, err := chain.GetTxSeens(p.Ctx, hashes)
	if err != nil {
		return fmt.Errorf("error getting tx seens; %w", err)
	}
	inputs, err := chain.GetTxInputsByHashes(p.Ctx, hashes)
	if err != nil {
		return fmt.Errorf("error getting tx inputs; %w", err)
	}
	outputs, err := chain.GetTxOutputsByHashes(p.Ctx, hashes)
	if err != nil {
		return fmt.Errorf("error getting tx outputs; %w", err)
	}
	var txMap = make(map[[32]byte]*chain.Tx, len(txs))
	for _, tx := range txs {
		txMap[tx.TxHash] = tx
	}
	var seenMap = make(map[[32]byte]*chain.TxSeen, len(seens))
	for _, seen := range seens {
		seenMap[seen.TxHash] = seen
	}
	var inMap = make(map[[32]byte][]*chain.TxInput)
	for _, input := range inputs {
		inMap[input.TxHash] = append(inMap[input.TxHash], input)
	}
	var outMap = make(map[[32]byte][]*chain.TxOutput)
	for _, output := range outputs {
		outMap[output.TxHash] = append(outMap[output.TxHash], output)
	}
	var block = &dbi.Block{}
	for _, txHash := range hashes {
		tx, ok := txMap[txHash]
		seen, seenOk := seenMap[txHash]
		if !ok || !seenOk || len(outMap[txHash]) == 0 {
			result.Missing = append(result.Missing, txHash)
			continue
		}
		msgTx := buildMsgTx(tx, inMap[txHash], outMap[txHash])
		if msgTx.TxHash() != chainhash.Hash(txHash) {
			// Rebuilt hash mismatch means the Index data for this tx is incomplete.
			result.Missing = append(result.Missing, txHash)
			continue
		}
		block.Transactions = append(block.Transactions, dbi.Tx{
			Hash:  txHash,
			Seen:  seen.Timestamp,
			MsgTx: msgTx,
		})
		result.Processed++
	}
	if len(block.Transactions) == 0 {
		return nil
	}
	if err := opReturnSaver.SaveTxs(p.Ctx, block); err != nil {
		return fmt.Errorf("error saving op returns; %w", err)
	}
	return nil
}

func buildMsgTx(tx *chain.Tx, inputs []*chain.TxInput, outputs []*chain.TxOutput) *wire.MsgTx {
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].Index < inputs[j].Index })
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].Index < outputs[j].Index })
	msgTx := wire.NewMsgTx(tx.Version)
	msgTx.LockTime = tx.LockTime
	for _, input := range inputs {
		outPoint := wire.NewOutPoint(chainHash(input.PrevHash), input.PrevIndex)
		txIn := wire.NewTxIn(outPoint, input.UnlockScript)
		txIn.Sequence = input.Sequence
		msgTx.AddTxIn(txIn)
	}
	for _, output := range outputs {
		msgTx.AddTxOut(wire.NewTxOut(output.Value, output.LockScript))
	}
	return msgTx
}

func chainHash(hash [32]byte) *chainhash.Hash {
	h := chainhash.Hash(hash)
	return &h
}
