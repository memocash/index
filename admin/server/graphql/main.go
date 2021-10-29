package graphql

import (
	_ "github.com/99designs/gqlgen/cmd"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/resolver"
	"net/http"
)

func GetHandler() (http.Handler, error) {
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{}})), nil
}
