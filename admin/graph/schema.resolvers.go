package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
)

func (r *mutationResolver) Null(ctx context.Context) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Tx(ctx context.Context, hash string) (*model.Tx, error) {
	chainHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, jerr.Get("error getting tx hash from hash", err)
	}
	txHash := chainHash.CloneBytes()
	txHashString := chainHash.String()
	preloads := GetPreloads(ctx)
	var raw []byte
	var msgTx *wire.MsgTx
	if jutil.StringsInSlice([]string{"raw", "inputs", "outputs"}, preloads) {
		txBlocks, err := item.GetSingleTxBlocks(txHash)
		if err != nil {
			return nil, jerr.Get("error getting tx blocks from items", err)
		}
		if len(txBlocks) == 0 {
			mempoolTxRaw, err := item.GetMempoolTxRawByHash(txHash)
			if err != nil {
				return nil, jerr.Get("error getting mempool tx raw", err)
			}
			raw = mempoolTxRaw.Raw
		} else {
			txRaw, err := item.GetRawBlockTxByHash(txBlocks[0].BlockHash, txHash)
			if err != nil {
				return nil, jerr.Get("error getting block tx by hash", err)
			}
			raw = txRaw.Raw
		}
		if jutil.StringsInSlice([]string{"inputs", "outputs"}, preloads) {
			msgTx, err = memo.GetMsgFromRaw(raw)
			if err != nil {
				return nil, jerr.Get("error getting tx msg from raw", err)
			}
		}
	}
	var inputs []*model.TxInput
	if jutil.StringInSlice("inputs", preloads) {
		var txOutputs []*item.TxOutput
		if jutil.StringInSlice("inputs.output", preloads) {
			var outs = make([]memo.Out, len(msgTx.TxIn))
			for i := range msgTx.TxIn {
				outs[i] = memo.Out{
					TxHash: msgTx.TxIn[i].PreviousOutPoint.Hash.CloneBytes(),
					Index:  msgTx.TxIn[i].PreviousOutPoint.Index,
				}
			}
			txOutputs, err = item.GetTxOutputs(outs)
			if err != nil {
				return nil, jerr.Get("error getting input tx outputs", err)
			}
		}
		for i, txIn := range msgTx.TxIn {
			var output *model.TxOutput
			for _, txOutput := range txOutputs {
				if bytes.Equal(txIn.PreviousOutPoint.Hash.CloneBytes(), txOutput.TxHash) &&
					txIn.PreviousOutPoint.Index == txOutput.Index {
					output = &model.TxOutput{
						Hash:   hs.GetTxString(txOutput.TxHash),
						Index:  txOutput.Index,
						Amount: txOutput.Value,
					}
				}
			}
			inputs = append(inputs, &model.TxInput{
				Hash:      txHashString,
				Index:     uint32(i),
				PrevHash:  txIn.PreviousOutPoint.Hash.String(),
				PrevIndex: txIn.PreviousOutPoint.Index,
				Output:    output,
			})
		}
	}
	var outputs []*model.TxOutput
	if jutil.StringInSlice("outputs", preloads) {
		for i, txOut := range msgTx.TxOut {
			outputs = append(outputs, &model.TxOutput{
				Hash:   txHashString,
				Index:  uint32(i),
				Amount: txOut.Value,
				Script: hex.EncodeToString(txOut.PkScript),
			})
		}
		if jutil.StringInSlice("outputs.spends", preloads) {
			var outs = make([]memo.Out, len(msgTx.TxOut))
			for i := range msgTx.TxOut {
				outs[i] = memo.Out{
					TxHash: txHash,
					Index:  uint32(i),
				}
			}
			outputInputs, err := item.GetOutputInputs(outs)
			if err != nil {
				return nil, jerr.Get("error getting output inputs for tx", err)
			}
			for _, outputInput := range outputInputs {
				if int(outputInput.PrevIndex) >= len(outputs) {
					return nil, jerr.Newf("error got output input out of range of outputs: %d %d", outputInput.PrevIndex, len(outputs))
				}
				outputInputHash, err := chainhash.NewHash(outputInput.Hash)
				if err != nil {
					return nil, jerr.Get("error getting output input hash", err)
				}
				outputs[outputInput.PrevIndex].Spends = append(outputs[outputInput.PrevIndex].Spends, &model.TxInput{
					Hash:      outputInputHash.String(),
					Index:     outputInput.Index,
					PrevHash:  txHashString,
					PrevIndex: outputInput.PrevIndex,
				})
			}
		}
	}
	return &model.Tx{
		Hash:    txHashString,
		Raw:     hex.EncodeToString(raw),
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}

func (r *queryResolver) Txs(ctx context.Context) ([]*model.Tx, error) {
	return []*model.Tx{{
		Hash: "123",
		Raw:  "456",
	}}, nil
}

func (r *queryResolver) TxInputs(ctx context.Context, hashIndexes []*model.HashIndex) ([]*model.TxInput, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) TxOutputs(ctx context.Context, hashIndexes []*model.HashIndex) ([]*model.TxOutput, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) TxInput(ctx context.Context, hash string, index uint32) (*model.TxInput, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) TxOutput(ctx context.Context, hash string, index uint32) (*model.TxOutput, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
