//go:generate go run github.com/99designs/gqlgen
package load

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/jchavannes/jgo/jutil"
	"log"
	"strings"
	"time"
)

const defaultWait = 10 * time.Millisecond

func HasField(fields []Field, check string) bool {
	return HasFieldAny(fields, []string{check})
}

func HasFieldAny(fields []Field, checks []string) bool {
	var fieldNames = make([]string, len(fields))
	for i, field := range fields {
		fieldNames[i] = field.Name
	}
	return jutil.StringsInSlice(checks, fieldNames)
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

func GetPrefixFields(fields []Field, prefix string) (prefixFields []Field) {
	for _, childField := range strings.Split(strings.TrimRight(prefix, "."), ".") {
		var foundFields []Field
		for _, field := range fields {
			if field.Name == childField {
				foundFields = field.Fields
			}
		}
		if len(foundFields) == 0 {
			return
		}
		fields = foundFields
	}
	prefixFields = fields
	return
}

type Field struct {
	Name      string
	Arguments map[string]interface{}
	Fields    []Field
}

func GetFields(ctx context.Context) []Field {
	return getFields(
		graphql.GetOperationContext(ctx),
		graphql.CollectFieldsCtx(ctx, nil),
	)
}

func getFields(ctx *graphql.OperationContext, fields []graphql.CollectedField) []Field {
	var fieldList []Field
	for _, field := range fields {
		var arguments = make(map[string]interface{})
		for _, arg := range field.Field.Arguments {
			arguments[arg.Name], _ = arg.Value.Value(ctx.Variables)
		}
		fieldList = append(fieldList, Field{
			Name:      field.Field.Name,
			Arguments: arguments,
			Fields:    getFields(ctx, graphql.CollectFields(ctx, field.Selections, nil)),
		})
	}
	return fieldList
}

func PrintFields(fields []Field, layer int) {
	spaces := strings.Repeat(" ", layer*2)
	for _, field := range fields {
		var args []string
		for name, val := range field.Arguments {
			args = append(args, fmt.Sprintf("%s: %v", name, val))
		}
		if len(args) > 0 {
			log.Printf("%s%s(%s)\n", spaces, field.Name, strings.Join(args, ", "))
		} else {
			log.Printf("%s%s\n", spaces, field.Name)
		}
		PrintFields(field.Fields, layer+1)
	}
}
