package lead

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"time"
)

type Counter struct {
	On        bool
	Counter   int
	StopChan  chan struct{}
	Ticker    *time.Ticker
	ErrorChan chan ShardError
}

func (c *Counter) Start(clients map[int]*Client, errorChan chan ShardError) {
	c.On = true
	c.Ticker = time.NewTicker(time.Second)
	c.StopChan = make(chan struct{})
	c.ErrorChan = errorChan
	jlog.Logf("Starting counter...\n")
	go func() {
		for {
		Select:
			select {
			case <-c.Ticker.C:
				for _, client := range clients {
					resp, err := client.Client.Ping(context.Background(), &cluster_pb.PingReq{
						Nonce: uint64(time.Now().UnixNano()),
					})
					if err != nil {
						c.ErrorChan <- ShardError{
							Shard: client.Config.Int(),
							Error: jerr.Get("error ping cluster shard", err),
						}
						break Select
					}
					jlog.Logf("Pinged shard %d, nonce: %d\n", client.Config.Shard, resp.Nonce)
				}
				c.Counter++
				jlog.Logf("Counter tick: %d\n", c.Counter)
				continue
			case <-c.StopChan:
			}
			jlog.Log("Stopping counter")
			c.Ticker.Stop()
			c.Ticker = nil
			c.On = false
			return
		}
	}()
}

func (c *Counter) Stop() {
	if c.On {
		close(c.StopChan)
	}
}
