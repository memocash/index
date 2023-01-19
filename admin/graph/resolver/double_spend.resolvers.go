package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

// Output is the resolver for the output field.
func (r *doubleSpendResolver) Output(ctx context.Context, obj *model.DoubleSpend) (*model.TxOutput, error) {
	if !HasFieldAny(ctx, []string{"amount", "script"}) {
		return &model.TxOutput{
			Hash:  obj.Hash,
			Index: obj.Index,
		}, nil
	}
	txOutput, err := dataloader.NewTxOutputLoader(txOutputLoaderConfig).Load(model.HashIndex{
		Hash:  obj.Hash,
		Index: obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting output for double spend from loader", err)
	}
	return txOutput, nil
}

// Inputs is the resolver for the inputs field.
func (r *doubleSpendResolver) Inputs(ctx context.Context, obj *model.DoubleSpend) ([]*model.TxInput, error) {
	hash, err := chainhash.NewHashFromStr(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error parsing double spend hash", err)
	}
	outputInputs, err := chain.GetOutputInput(memo.Out{
		TxHash: hash.CloneBytes(),
		Index:  obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting output inputs for double spend", err)
	}
	var txInputs = make([]*model.TxInput, len(outputInputs))
	for i := range outputInputs {
		txInputs[i] = &model.TxInput{
			Hash:      chainhash.Hash(outputInputs[i].Hash).String(),
			Index:     outputInputs[i].Index,
			PrevHash:  chainhash.Hash(outputInputs[i].PrevHash).String(),
			PrevIndex: outputInputs[i].PrevIndex,
		}
	}
	return txInputs, nil
}

// DoubleSpend returns generated.DoubleSpendResolver implementation.
func (r *Resolver) DoubleSpend() generated.DoubleSpendResolver { return &doubleSpendResolver{r} }

type doubleSpendResolver struct{ *Resolver }
