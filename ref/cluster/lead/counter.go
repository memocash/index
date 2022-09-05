package lead

import (
	"github.com/jchavannes/jgo/jlog"
	"time"
)

type Counter struct {
	On       bool
	Counter  int
	StopChan chan struct{}
	Ticker   *time.Ticker
}

func (c *Counter) Start() {
	c.On = true
	c.Ticker = time.NewTicker(time.Second)
	c.StopChan = make(chan struct{})
	go func() {
		for {
			select {
			case <-c.Ticker.C:
				c.Counter++
				jlog.Logf("Counter tick: %d\n", c.Counter)
			case <-c.StopChan:
				jlog.Log("Stopping counter")
				c.Ticker.Stop()
				c.Ticker = nil
				c.On = false
				return
			}
		}
	}()
}

func (c *Counter) Stop() {
	if c.On {
		close(c.StopChan)
	}
}
