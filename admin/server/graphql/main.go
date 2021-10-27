package graphql

import (
	_ "github.com/99designs/gqlgen/cmd"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/admin/graph"
	"github.com/memocash/server/admin/graph/generated"
	"net/http"
)

var graphqlRoute = admin.Route{
	Pattern: admin.UrlGraphql,
	Handler: nil,
}

func GetRoutes() []admin.Route {
	return []admin.Route{
		graphqlRoute,
	}
}

func GetHandler() (http.Handler, error) {
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})), nil
}

/*func GetHandler() (http.Handler, error) {
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, jerr.Get("error getting graphql schema", err)
	}
	return handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	}), nil
}*/
