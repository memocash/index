package load

import (
	"context"
	"sync"
)

type baseA struct {
	Ctx    context.Context
	Fields Fields
	Mutex  sync.Mutex
	mutexB sync.Mutex
	Wait   sync.WaitGroup
	Errors []error
}

func (b *baseA) HasField(checks []string) bool {
	b.mutexB.Lock()
	defer b.mutexB.Unlock()
	return b.Fields.HasFieldAny(checks)
}

func (b *baseA) AddError(err error) {
	b.mutexB.Lock()
	defer b.mutexB.Unlock()
	b.Errors = append(b.Errors, err)
}
