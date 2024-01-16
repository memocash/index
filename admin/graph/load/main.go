//go:generate go run github.com/99designs/gqlgen
package load

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"log"
	"strings"
	"time"
)

const defaultWait = 10 * time.Millisecond

type Fields []Field

func (f Fields) HasField(check string) bool {
	return f.HasFieldAny([]string{check})
}

func (f Fields) HasFieldAny(checks []string) bool {
	for _, check := range checks {
		var checkFields = f
		var found bool
		for _, childCheck := range strings.Split(check, ".") {
			found = false
			for _, checkField := range checkFields {
				if checkField.Name == childCheck {
					found = true
					checkFields = checkField.Fields
					break
				}
			}
			if found == false {
				break
			}
		}
		if found == true {
			return true
		}
	}
	return false
}

func (f Fields) Print(layer int) {
	PrintFields(f, layer)
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

func GetFields(ctx context.Context) Fields {
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
