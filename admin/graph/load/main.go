//go:generate go run github.com/99designs/gqlgen
package load

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/jchavannes/jgo/jutil"
	"time"
)

const defaultWait = 10 * time.Millisecond

func HasField(ctx context.Context, preload string) bool {
	return jutil.StringInSlice(preload, GetPreloads(ctx))
}

func HasFieldAny(ctx context.Context, preloads []string) bool {
	return jutil.StringsInSlice(preloads, GetPreloads(ctx))
}

func GetPreloads(ctx context.Context) []string {
	return GetNestedPreloads(
		graphql.GetOperationContext(ctx),
		graphql.CollectFieldsCtx(ctx, nil),
		"",
	)
}

func GetNestedPreloads(ctx *graphql.OperationContext, fields []graphql.CollectedField, prefix string) (preloads []string) {
	for _, column := range fields {
		prefixColumn := GetPreloadString(prefix, column.Name)
		preloads = append(preloads, prefixColumn)
		preloads = append(preloads, GetNestedPreloads(ctx, graphql.CollectFields(ctx, column.Selections, nil), prefixColumn)...)
	}
	return
}

func GetPreloadString(prefix, name string) string {
	if len(prefix) > 0 {
		return prefix + "." + name
	}
	return name
}
