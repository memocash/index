package lead

import (
	"context"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"sync"
)

type Processor struct {
	On        bool
	StopChan  chan struct{}
	Clients   map[int]*Client
	ErrorChan chan ShardError
	Node      *Node
}

func (p *Processor) Start() {
	if p.On {
		return
	}
	p.StopChan = make(chan struct{})
	p.On = true
	p.Node = NewNode()
	p.Node.Start()
	jlog.Logf("Starting node listener...\n")
	go func() {
		for {
			select {
			case block := <-p.Node.NewBlock:
				if p.Process(block) {
					continue
				}
			case <-p.StopChan:
			}
			jlog.Log("Stopping node listener")
			return
		}
	}()
}

func (p *Processor) Process(block *wire.MsgBlock) bool {
	if !p.On {
		return false
	}
	var shardBlocks = make(map[uint32]*wire.MsgBlock)
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		shard := db.GetShardByte32(txHash[:])
		if _, ok := shardBlocks[shard]; !ok {
			shardBlocks[shard] = wire.NewMsgBlock(&block.Header)
		}
		shardBlocks[shard].AddTransaction(tx)
	}
	blockHash := block.BlockHash()
	var wg sync.WaitGroup
	var hadError bool
	for _, client := range p.Clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			if _, ok := shardBlocks[client.Config.Shard]; !ok {
				return
			}
			if _, err := client.Client.Process(context.Background(), &cluster_pb.ProcessReq{
				Block: memo.GetRawBlock(*shardBlocks[client.Config.Shard]),
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
		jlog.Logf("Processed block: %s %s, %d txs\n", blockHash, block.Header.Timestamp, len(block.Transactions))
	}
	return true
}

func (p *Processor) Stop() {
	if p.On {
		p.On = false
		close(p.StopChan)
		p.Node.Stop()
	}
}

func NewProcessor(clients map[int]*Client, errorChan chan ShardError) *Processor {
	return &Processor{
		Clients:   clients,
		ErrorChan: errorChan,
	}
}
