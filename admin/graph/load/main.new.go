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
	TxInputLoader *dataloader.Loader
}

func NewLoaders() *Loaders {
	txInputReader := &TxInputReader{}
	loaders := &Loaders{
		TxInputLoader: dataloader.NewBatchedLoader(txInputReader.GetTxInputs),
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
