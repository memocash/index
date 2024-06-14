package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	_ "github.com/99designs/gqlgen/graphql/introspection"
	"github.com/gorilla/websocket"
	"github.com/memocash/index/graph/generated"
	"github.com/memocash/index/graph/resolver"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"net"
	"net/http"
	"time"
)

func getGqlGenHandler() *handler.Server {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{}}))
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.SetQueryCache(lru.New(1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})
	srv.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		pathStr := graphql.GetPath(ctx).String()
		if pathStr != "" {
			pathStr = " (" + pathStr + ")"
		}
		resolver.LogContextRequest(ctx, fmt.Sprintf("error processing request%s; %v", pathStr, e))
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
	return srv
}

func GetGraphQLHandler() func(w http.ResponseWriter, r *http.Request) {
	srv := getGqlGenHandler()
	return func(w http.ResponseWriter, r *http.Request) {
		graphRequest := resolver.NewRequest(getIpAddress(r), "/graphql")
		var finalMessages []string
		var writer *sizeWriter
		if r.Header.Get("Upgrade") != "" {
			finalMessages = append(finalMessages, "[close]")
		} else {
			writer = &sizeWriter{httpWriter: w}
			w = writer
		}
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Server")
		r = r.WithContext(resolver.AttachRequestToContext(r.Context(), graphRequest))
		srv.ServeHTTP(w, r)
		if writer != nil {
			graphRequest.Size = writer.totalSize
		}
		graphRequest.LogFinal(finalMessages...)
	}
}

func getIpAddress(r *http.Request) string {
	cfIp := r.Header.Get("CF-Connecting-IP")
	if cfIp != "" {
		return cfIp
	}
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	remoteHost, _, _ := net.SplitHostPort(r.RemoteAddr)
	return remoteHost
}
