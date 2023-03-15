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

type TxInputReader struct {
}

func (r *TxInputReader) GetTxInputs(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	var results = make([]*dataloader.Result, len(keys))
	var txHashes [][32]byte
	for i := range keys {
		txHash, err := chainhash.NewHashFromStr(keys[i].String())
		if err != nil {
			results[i] = &dataloader.Result{Error: fmt.Errorf("error getting tx hash for tx inputs dataloader; %w", err)}
			continue
		}
		txHashes = append(txHashes, *txHash)
	}
	txInputs, err := chain.GetTxInputsByHashes(txHashes)
	if err != nil {
		for i := range results {
			if results[i] == nil {
				results[i] = &dataloader.Result{Error: fmt.Errorf("error getting tx inputs for dataloader; %w", err)}
			}
		}
		return results
	}
	var txInputsByTxHash = make(map[string][]*model.TxInput)
	for _, txInput := range txInputs {
		txHash := chainhash.Hash(txInput.TxHash).String()
		txInputsByTxHash[txHash] = append(txInputsByTxHash[txHash], &model.TxInput{
			Hash:      chainhash.Hash(txInput.TxHash).String(),
			Index:     txInput.Index,
			PrevHash:  chainhash.Hash(txInput.PrevHash).String(),
			PrevIndex: txInput.PrevIndex,
			Script:    hex.EncodeToString(txInput.UnlockScript),
		})
	}
	for index, txHash := range keys {
		hashTxInputs, ok := txInputsByTxHash[txHash.String()]
		if ok {
			results[index] = &dataloader.Result{Data: hashTxInputs}
		} else if results[index] == nil {
			results[index] = &dataloader.Result{Error: fmt.Errorf("tx input not found in dataloader %s", txHash)}
		}
	}
	return results
}

func GetTxInputs(ctx context.Context, txHash string) ([]*model.TxInput, error) {
	loaders := For(ctx)
	thunk := loaders.TxInputLoader.Load(ctx, dataloader.StringKey(txHash))
	result, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("error getting tx inputs from loader; %w", err)
	}
	return result.([]*model.TxInput), nil
}
