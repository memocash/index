package maint

import (
	"fmt"
	"sync"
)

type ShardProcess struct {
	Errors []error
	Mutex  sync.Mutex
	Wg     sync.WaitGroup
}

func (p *ShardProcess) AddError(shard uint32, err error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.Errors = append(p.Errors, fmt.Errorf("error process shard: %d; %w", shard, err))
}
