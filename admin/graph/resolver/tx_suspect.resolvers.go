package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
)

func (r *txSuspectResolver) Tx(ctx context.Context, obj *model.TxSuspect) (*model.Tx, error) {
	tx, err := TxLoader(ctx, obj.Hash)
	if err != nil {
		return nil, jerr.Get("error getting tx from loader for tx suspect resolver", err)
	}
	return tx, nil
}

// TxSuspect returns generated.TxSuspectResolver implementation.
func (r *Resolver) TxSuspect() generated.TxSuspectResolver { return &txSuspectResolver{r} }

type txSuspectResolver struct{ *Resolver }
