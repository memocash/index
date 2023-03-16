package load

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/graph-gophers/dataloader"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

func HashInputString(txHash string, index uint32) string {
	return fmt.Sprintf("%s %d", txHash, index)
}

func HashInputFromString(input string) (string, uint32, error) {
	var txHash string
	var index uint32
	if _, err := fmt.Sscanf(input, "%s %d", &txHash, &index); err != nil {
		return "", 0, fmt.Errorf("error getting tx hash and index from string; %w", err)
	}
	return txHash, index, nil
}

type OutputInputsReader struct {
}

func (r *OutputInputsReader) GetOutputInput(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	return getTxOutputInputs(ctx, keys, false)
}

func GetOutputInputs(ctx context.Context, txHash string, index uint32) ([]*model.TxInput, error) {
	loaders := For(ctx)
	thunk := loaders.OutputInputsLoader.Load(ctx, dataloader.StringKey(HashInputString(txHash, index)))
	result, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("error getting tx output inputs from loader; %w", err)
	}
	return result.([]*model.TxInput), nil
}

type OutputInputsWithScriptReader struct {
}

func (r *OutputInputsWithScriptReader) GetOutputInput(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	return getTxOutputInputs(ctx, keys, true)
}

func GetOutputInputsWithScript(ctx context.Context, txHash string, index uint32) ([]*model.TxInput, error) {
	loaders := For(ctx)
	thunk := loaders.OutputInputsWithScriptLoader.Load(ctx, dataloader.StringKey(HashInputString(txHash, index)))
	result, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("error getting tx output inputs with script from loader; %w", err)
	}
	return result.([]*model.TxInput), nil
}

func getTxOutputInputs(ctx context.Context, keys dataloader.Keys, withScript bool) []*dataloader.Result {
	var results = make([]*dataloader.Result, len(keys))
	var outs []memo.Out
	for i := range keys {
		txHash, index, err := HashInputFromString(keys[i].String())
		if err != nil {
			results[i] = &dataloader.Result{
				Error: fmt.Errorf("error getting tx hash index for tx output inputs dataloader: %s; %w", keys[i], err)}
			continue
		}
		hash, err := chainhash.NewHashFromStr(txHash)
		if err != nil {
			results[i] = &dataloader.Result{
				Error: fmt.Errorf("error getting tx hash for tx output inputs dataloader; %w", err)}
			continue
		}
		outs = append(outs, memo.Out{
			TxHash: hash[:],
			Index:  index,
		})
	}
	outputInputs, err := chain.GetOutputInputs(outs)
	if err != nil {
		return resultsError(results, fmt.Errorf("error getting tx output inputs from chain; %w", err))
	}
	var txInputs []*chain.TxInput
	if withScript {
		var ins = make([]memo.Out, len(outputInputs))
		for i := range outputInputs {
			ins[i] = memo.Out{
				TxHash: outputInputs[i].Hash[:],
				Index:  outputInputs[i].Index,
			}
		}
		if txInputs, err = chain.GetTxInputs(ins); err != nil {
			return resultsError(results, fmt.Errorf("error getting tx output inputs script from chain; %w", err))
		}
	}
	var outputInputsByTxHashIndex = make(map[string][]*model.TxInput)
	for _, outputInput := range outputInputs {
		prevHash := chainhash.Hash(outputInput.PrevHash).String()
		var modelOutputInput = &model.TxInput{
			Hash:      chainhash.Hash(outputInput.Hash).String(),
			Index:     outputInput.Index,
			PrevHash:  prevHash,
			PrevIndex: outputInput.PrevIndex,
		}
		for _, txInput := range txInputs {
			if txInput.TxHash == outputInput.Hash && txInput.Index == outputInput.Index {
				modelOutputInput.Script = hex.EncodeToString(txInput.UnlockScript)
				break
			}
		}
		hashIndex := HashInputString(prevHash, outputInput.PrevIndex)
		outputInputsByTxHashIndex[hashIndex] = append(outputInputsByTxHashIndex[hashIndex], modelOutputInput)
	}
	for index, hashIndex := range keys {
		hashIndexTxInput, ok := outputInputsByTxHashIndex[hashIndex.String()]
		if ok {
			results[index] = &dataloader.Result{Data: hashIndexTxInput}
		} else if results[index] == nil {
			results[index] = &dataloader.Result{Data: []*model.TxInput{}}
		}
	}
	return results
}
