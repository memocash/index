//go:build debug

package server

import (
	"github.com/99designs/gqlgen/graphql/playground"
	"net/http"
)

func init() {
	AddHandlers = append(AddHandlers, func(mux *http.ServeMux) {
		mux.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
		mux.Handle("/playground-live", playground.Handler("GraphQL playground", "https://graph.cash/graphql"))
	})
}
