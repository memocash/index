package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *txInputResolver) Tx(ctx context.Context, obj *model.TxInput) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *txInputResolver) Output(ctx context.Context, obj *model.TxInput) (*model.TxOutput, error) {
	/*hash, err := chainhash.NewHashFromStr(obj.PrevHash)
	if err != nil {
		return nil, jerr.Get("error parsing input tx hash for output", err)
	}
	txOutput, err := item.GetTxOutput(hash.CloneBytes(), obj.PrevIndex)
	if err != nil {
		return nil, jerr.Get("error getting input tx outputs", err)
	}
	return &model.TxOutput{
		Hash:   hs.GetTxString(txOutput.TxHash),
		Index:  txOutput.Index,
		Amount: txOutput.Value,
	}, nil*/
	return &model.TxOutput{
		Hash:  obj.PrevHash,
		Index: obj.PrevIndex,
	}, nil
}

// TxInput returns generated.TxInputResolver implementation.
func (r *Resolver) TxInput() generated.TxInputResolver { return &txInputResolver{r} }

type txInputResolver struct{ *Resolver }
