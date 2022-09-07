package lead

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"sync"
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
	if c.On {
		return
	}
	c.StopChan = make(chan struct{})
	c.On = true
	c.Ticker = time.NewTicker(time.Second)
	c.ErrorChan = errorChan
	jlog.Logf("Starting counter...\n")
	go func() {
		for {
		Select:
			select {
			case <-c.Ticker.C:
				var wg sync.WaitGroup
				var hadError bool
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
					wg.Add(1)
					go func(client *Client) {
						defer wg.Done()
						if _, err := client.Client.Process(context.Background(), &cluster_pb.ProcessReq{
							Block: jutil.GetIntData(c.Counter),
						}); err != nil {
							hadError = true
							c.ErrorChan <- ShardError{
								Shard: client.Config.Int(),
								Error: jerr.Getf(err, "error cluster shard process: %d", client.Config.Shard),
							}
						}
					}(client)
				}
				wg.Wait()
				if !hadError {
					c.Counter++
					jlog.Logf("Counter tick: %d\n", c.Counter)
				}
				continue
			case <-c.StopChan:
			}
			jlog.Log("Stopping counter")
			c.Ticker.Stop()
			c.Ticker = nil
			return
		}
	}()
}

func (c *Counter) Stop() {
	if c.On {
		c.On = false
		close(c.StopChan)
	}
}
