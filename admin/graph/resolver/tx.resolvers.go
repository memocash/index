package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/graph/dataloader"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/ref/bitcoin/memo"
)

func (r *txResolver) Inputs(ctx context.Context, obj *model.Tx) ([]*model.TxInput, error) {
	rawBytes, err := hex.DecodeString(obj.Raw)
	if err != nil {
		return nil, jerr.Get("error decoding raw tx", err)
	}
	msgTx, err := memo.GetMsgFromRaw(rawBytes)
	if err != nil {
		return nil, jerr.Get("error getting tx msg from raw", err)
	}
	var inputs = make([]*model.TxInput, len(msgTx.TxIn))
	for i := range msgTx.TxIn {
		inputs[i] = &model.TxInput{
			Hash:      obj.Hash,
			Index:     uint32(i),
			PrevHash:  msgTx.TxIn[i].PreviousOutPoint.Hash.String(),
			PrevIndex: msgTx.TxIn[i].PreviousOutPoint.Index,
		}
	}
	return inputs, nil
}

func (r *txResolver) Outputs(ctx context.Context, obj *model.Tx) ([]*model.TxOutput, error) {
	rawBytes, err := hex.DecodeString(obj.Raw)
	if err != nil {
		return nil, jerr.Get("error decoding raw tx", err)
	}
	msgTx, err := memo.GetMsgFromRaw(rawBytes)
	if err != nil {
		return nil, jerr.Get("error getting tx msg from raw", err)
	}
	var outputs = make([]*model.TxOutput, len(msgTx.TxOut))
	for i := range msgTx.TxOut {
		outputs[i] = &model.TxOutput{
			Hash:   obj.Hash,
			Index:  uint32(i),
			Amount: msgTx.TxOut[i].Value,
			Script: hex.EncodeToString(msgTx.TxOut[i].PkScript),
		}
	}
	return outputs, nil
}

func (r *txResolver) Blocks(ctx context.Context, obj *model.Tx) ([]*model.Block, error) {
	blocks, err := dataloader.NewBlockLoader(blockLoaderConfig).Load(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error getting blocks for tx from loader", err)
	}
	return blocks, nil

}

func (r *txResolver) Suspect(ctx context.Context, obj *model.Tx) (*model.TxSuspect, error) {
	txSuspect, err := dataloader.NewTxSuspectLoader(txSuspectLoaderConfig).Load(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error getting tx suspect for tx from loader", err)
	}
	return txSuspect, nil
}

func (r *txResolver) Lost(ctx context.Context, obj *model.Tx) (*model.TxLost, error) {
	txLost, err := dataloader.NewTxLostLoader(txLostLoaderConfig).Load(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error getting tx lost for tx from loader", err)
	}
	return txLost, nil
}

// Tx returns generated.TxResolver implementation.
func (r *Resolver) Tx() generated.TxResolver { return &txResolver{r} }

type txResolver struct{ *Resolver }
