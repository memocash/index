package load

import (
	"context"
	"github.com/graph-gophers/dataloader"
	"net/http"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

type Loaders struct {
	TxBlocksLoader               *dataloader.Loader
	TxBlocksWithInfoLoader       *dataloader.Loader
	TxInputsLoader               *dataloader.Loader
	TxOutputsLoader              *dataloader.Loader
	OutputInputsLoader           *dataloader.Loader
	OutputInputsWithScriptLoader *dataloader.Loader
}

func NewLoaders() *Loaders {
	var (
		txBlocksReader               = &TxBlocksReader{}
		txBlocksWithInfoReader       = &TxBlocksWithInfoReader{}
		txInputsReader               = &TxInputsReader{}
		txOutputsReader              = &TxOutputsReader{}
		outputInputsReader           = &OutputInputsReader{}
		outputInputsWithScriptReader = &OutputInputsWithScriptReader{}
	)
	loaders := &Loaders{
		TxBlocksLoader:               dataloader.NewBatchedLoader(txBlocksReader.GetTxBlocks),
		TxBlocksWithInfoLoader:       dataloader.NewBatchedLoader(txBlocksWithInfoReader.GetTxBlocks),
		TxInputsLoader:               dataloader.NewBatchedLoader(txInputsReader.GetTxInputs),
		TxOutputsLoader:              dataloader.NewBatchedLoader(txOutputsReader.GetTxOutputs),
		OutputInputsLoader:           dataloader.NewBatchedLoader(outputInputsReader.GetOutputInput),
		OutputInputsWithScriptLoader: dataloader.NewBatchedLoader(outputInputsWithScriptReader.GetOutputInput),
	}
	return loaders
}

func Middleware(loaders *Loaders, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCtx := context.WithValue(r.Context(), loadersKey, loaders)
		r = r.WithContext(nextCtx)
		next.ServeHTTP(w, r)
	})
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

func resultsError(results []*dataloader.Result, err error) []*dataloader.Result {
	for i := range results {
		if results[i] == nil {
			results[i] = &dataloader.Result{Error: err}
		}
	}
	return results
}
