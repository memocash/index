package server

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	_ "github.com/99designs/gqlgen/graphql/introspection"
	"github.com/memocash/index/graph/generated"
	"github.com/memocash/index/graph/resolver"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"log"
	"net/http"
)

func GetGraphQLHandler() func(w http.ResponseWriter, r *http.Request) {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{}}))
	srv.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		log.Printf("error processing request (%s); %v", graphql.GetPath(ctx), e)
		var internalError resolver.InternalError
		if errors.As(e, &internalError) {
			return &gqlerror.Error{
				Message: "Internal server error",
				Extensions: map[string]interface{}{
					"code": "INTERNAL_SERVER_ERROR",
				},
			}
		}
		return graphql.DefaultErrorPresenter(ctx, e)
	})
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Server")
		srv.ServeHTTP(w, r)
		log.Printf("Processed graph request: /graphql\n")
	}
}
