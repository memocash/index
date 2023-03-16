package load

import (
	"github.com/jchavannes/jgo/jutil"
	"sync"
)

type baseA struct {
	Preloads []string
	Mutex    sync.Mutex
	Wait     sync.WaitGroup
	Errors   []error
}

func (b *baseA) HasPreload(check []string) bool {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	return jutil.StringsInSlice(check, b.Preloads)
}

func (b *baseA) AddError(err error) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.Errors = append(b.Errors, err)
}
