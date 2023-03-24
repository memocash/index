package load

import (
	"context"
	"github.com/jchavannes/jgo/jutil"
	"sync"
)

type baseA struct {
	Ctx      context.Context
	Preloads []string
	Mutex    sync.Mutex
	mutexB   sync.Mutex
	Wait     sync.WaitGroup
	Errors   []error
}

func (b *baseA) HasPreload(check []string) bool {
	b.mutexB.Lock()
	defer b.mutexB.Unlock()
	return jutil.StringsInSlice(check, b.Preloads)
}

func (b *baseA) AddError(err error) {
	b.mutexB.Lock()
	defer b.mutexB.Unlock()
	b.Errors = append(b.Errors, err)
}
