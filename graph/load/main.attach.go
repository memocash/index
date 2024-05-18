package load

import (
	"context"
	"sync"
	"time"
)

type baseA struct {
	Ctx    context.Context
	Fields Fields
	Mutex  mutex1second
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
