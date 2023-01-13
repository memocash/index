package maint

import (
	"github.com/jchavannes/jgo/jerr"
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
	p.Errors = append(p.Errors, jerr.Getf(err, "error process shard: %d", shard))
}
