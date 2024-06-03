package client

import (
	"log"
	"sync"
	"time"
)

type stat struct {
	id    int
	count int
}

func (s *stat) incr() {
	_stats.mutex.Lock()
	defer _stats.mutex.Unlock()
	s.count++
}

func removeStat(s *stat) {
	_stats.mutex.Lock()
	defer _stats.mutex.Unlock()
	delete(_stats.stats, s.id)
}

type stats struct {
	incr    int
	stats   map[int]*stat
	mutex   sync.Mutex
	tick    *time.Ticker
	lastNum int
}

var _stats = stats{}

func newStat() *stat {
	_stats.mutex.Lock()
	defer _stats.mutex.Unlock()
	_stats.incr++
	_stats.stats[_stats.incr] = &stat{id: _stats.incr}
	return _stats.stats[_stats.incr]
}

func startStats() {
	_stats.stats = make(map[int]*stat)
	_stats.tick = time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-_stats.tick.C:
				_stats.mutex.Lock()
				var totalCount int
				for _, s := range _stats.stats {
					totalCount += s.count
					s.count = 0
				}
				var changeNumStats bool
				var numStats = len(_stats.stats)
				if _stats.lastNum != numStats {
					_stats.lastNum = numStats
					changeNumStats = true
				}
				_stats.mutex.Unlock()
				if totalCount > 0 || changeNumStats {
					log.Printf("Subscriptions: %d, Messages: %d\n", numStats, totalCount)
				}
			}
		}
	}()
}
