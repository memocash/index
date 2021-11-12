package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/graph/dataloader"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/node/obj/get"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

func (r *txOutputResolver) Tx(ctx context.Context, obj *model.TxOutput) (*model.Tx, error) {
	return &model.Tx{
		Hash: obj.Hash,
	}, nil
}

func (r *txOutputResolver) Spends(ctx context.Context, obj *model.TxOutput) ([]*model.TxInput, error) {
	txInputs, err := dataloader.NewTxOutputSpendLoader(txOutputSpendLoaderConfig).Load(model.HashIndex{
		Hash:  obj.Hash,
		Index: obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for spends from loader", err)
	}
	return txInputs, nil
}

func (r *txOutputResolver) DoubleSpend(ctx context.Context, obj *model.TxOutput) (*model.DoubleSpend, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *txOutputResolver) Lock(ctx context.Context, obj *model.TxOutput) (*model.Lock, error) {
	if len(obj.Script) == 0 {
		return nil, nil
	}
	lockScript, err := hex.DecodeString(obj.Script)
	if err != nil {
		return nil, jerr.Get("error parsing lock script for tx output lock resolver", err)
	}
	address, err := wallet.GetAddressFromPkScript(lockScript)
	if err != nil {
		return nil, jerr.Get("error getting address from lock script", err)
	}
	balance := get.NewBalance(lockScript)
	if err := balance.GetBalance(); err != nil {
		return nil, jerr.Get("error getting lock balance for tx output resolver", err)
	}
	return &model.Lock{
		Hash:    hex.EncodeToString(script.GetLockHash(lockScript)),
		Address: address.GetEncoded(),
		Balance: balance.Balance,
	}, nil
}

// TxOutput returns generated.TxOutputResolver implementation.
func (r *Resolver) TxOutput() generated.TxOutputResolver { return &txOutputResolver{r} }

type txOutputResolver struct{ *Resolver }
