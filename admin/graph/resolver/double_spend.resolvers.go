package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

func (r *doubleSpendResolver) Output(ctx context.Context, obj *model.DoubleSpend) (*model.TxOutput, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *doubleSpendResolver) Inputs(ctx context.Context, obj *model.DoubleSpend) ([]*model.TxInput, error) {
	hash, err := chainhash.NewHashFromStr(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error parsing double spend hash", err)
	}
	outputInputs, err := item.GetOutputInput(memo.Out{
		TxHash: hash.CloneBytes(),
		Index:  obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting output inputs for double spend", err)
	}
	var txInputs = make([]*model.TxInput, len(outputInputs))
	for i := range outputInputs {
		txInputs[i] = &model.TxInput{
			Hash:      hs.GetTxString(outputInputs[i].Hash),
			Index:     outputInputs[i].Index,
			PrevHash:  hs.GetTxString(outputInputs[i].PrevHash),
			PrevIndex: outputInputs[i].PrevIndex,
		}
	}
	return txInputs, nil
}

// DoubleSpend returns generated.DoubleSpendResolver implementation.
func (r *Resolver) DoubleSpend() generated.DoubleSpendResolver { return &doubleSpendResolver{r} }

type doubleSpendResolver struct{ *Resolver }
