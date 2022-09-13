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

type Processor struct {
	On        bool
	Counter   int
	StopChan  chan struct{}
	Ticker    *time.Ticker
	Clients   map[int]*Client
	ErrorChan chan ShardError
}

func (p *Processor) Start() {
	if p.On {
		return
	}
	p.StopChan = make(chan struct{})
	p.On = true
	jlog.Logf("Starting counter...\n")
	go func() {
		for {
			select {
			case <-time.NewTimer(time.Second).C:
				if p.Process() {
					continue
				}
			case <-p.StopChan:
			}
			jlog.Log("Stopping counter")
			return
		}
	}()
}

func (p *Processor) Process() bool {
	if !p.On {
		return false
	}
	var wg sync.WaitGroup
	var hadError bool
	for _, client := range p.Clients {
		resp, err := client.Client.Ping(context.Background(), &cluster_pb.PingReq{
			Nonce: uint64(time.Now().UnixNano()),
		})
		if err != nil {
			p.ErrorChan <- ShardError{
				Shard: client.Config.Int(),
				Error: jerr.Get("error ping cluster shard", err),
			}
			return false
		}
		jlog.Logf("Pinged shard %d, nonce: %d\n", client.Config.Shard, resp.Nonce)
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			if _, err := client.Client.Process(context.Background(), &cluster_pb.ProcessReq{
				Block: jutil.GetIntData(p.Counter),
			}); err != nil {
				hadError = true
				p.ErrorChan <- ShardError{
					Shard: client.Config.Int(),
					Error: jerr.Getf(err, "error cluster shard process: %d", client.Config.Shard),
				}
			}
		}(client)
	}
	wg.Wait()
	if !hadError {
		p.Counter++
		jlog.Logf("Processor tick: %d\n", p.Counter)
	}
	return true
}

func (p *Processor) Stop() {
	if p.On {
		p.On = false
		close(p.StopChan)
	}
}

func NewProcessor(clients map[int]*Client, errorChan chan ShardError) *Processor {
	return &Processor{
		Clients:   clients,
		ErrorChan: errorChan,
	}
}
