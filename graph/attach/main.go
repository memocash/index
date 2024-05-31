package attach

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/graph/model"
	"log"
	"strings"
	"sync"
	"time"
)

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

func (f Fields) GetField(check string) Field {
	for _, field := range f {
		if field.Name == check {
			return field
		}
	}
	return Field{}
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
	Fields    Fields
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

type base struct {
	Ctx    context.Context
	Fields Fields
	Mutex  mutex1second
	mutexB sync.Mutex
	Wait   sync.WaitGroup
	Errors []error
}

func (b *base) HasField(checks []string) bool {
	b.mutexB.Lock()
	defer b.mutexB.Unlock()
	return b.Fields.HasFieldAny(checks)
}

func (b *base) AddError(err error) {
	b.mutexB.Lock()
	defer b.mutexB.Unlock()
	b.Errors = append(b.Errors, err)
}

type mutex1second struct {
	mutex sync.Mutex
}

func (m *mutex1second) Lock() {
	var locked = make(chan struct{})
	go func() {
		m.mutex.Lock()
		close(locked)
	}()
	t := time.NewTimer(time.Second)
	select {
	case <-locked:
		t.Stop()
	case <-t.C:
		panic("mutex held locked for more than 1 second")
	}
}

func (m *mutex1second) Unlock() {
	m.mutex.Unlock()
}

func SetNameExists(setName *model.SetName) bool {
	return setName != nil && !jutil.AllZeros(setName.TxHash[:])
}

func SetProfileExists(setProfile *model.SetProfile) bool {
	return setProfile != nil && !jutil.AllZeros(setProfile.TxHash[:])
}

func SetPicExists(setPic *model.SetPic) bool {
	return setPic != nil && !jutil.AllZeros(setPic.TxHash[:])
}
