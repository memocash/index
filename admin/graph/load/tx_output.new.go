package load

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/graph-gophers/dataloader"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
)

type TxOutputsReader struct {
}

func (r *TxOutputsReader) GetTxOutputs(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	var results = make([]*dataloader.Result, len(keys))
	var txHashes [][32]byte
	for i := range keys {
		txHash, err := chainhash.NewHashFromStr(keys[i].String())
		if err != nil {
			results[i] = &dataloader.Result{Error: fmt.Errorf("error getting tx hash for tx outputs dataloader; %w", err)}
			continue
		}
		txHashes = append(txHashes, *txHash)
	}
	txOutputs, err := chain.GetTxOutputsByHashes(txHashes)
	if err != nil {
		return resultsError(results, fmt.Errorf("error getting tx outputs for dataloader; %w", err))
	}
	var txOutputsByTxHash = make(map[string][]*model.TxOutput)
	for _, txOutput := range txOutputs {
		txHash := chainhash.Hash(txOutput.TxHash).String()
		txOutputsByTxHash[txHash] = append(txOutputsByTxHash[txHash], &model.TxOutput{
			Hash:   chainhash.Hash(txOutput.TxHash).String(),
			Index:  txOutput.Index,
			Script: hex.EncodeToString(txOutput.LockScript),
			Amount: txOutput.Value,
		})
	}
	for index, txHash := range keys {
		hashTxOutputs, ok := txOutputsByTxHash[txHash.String()]
		if ok {
			results[index] = &dataloader.Result{Data: hashTxOutputs}
		} else if results[index] == nil {
			results[index] = &dataloader.Result{Error: fmt.Errorf("tx output not found in dataloader %s", txHash)}
		}
	}
	return results
}

func GetTxOutputs(ctx context.Context, txHash string) ([]*model.TxOutput, error) {
	loaders := For(ctx)
	thunk := loaders.TxOutputsLoader.Load(ctx, dataloader.StringKey(txHash))
	result, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("error getting tx outputs from loader; %w", err)
	}
	return result.([]*model.TxOutput), nil
}
