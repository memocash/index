package lead

import (
	"context"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/obj/saver"
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
	Verbose   bool
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
	if !p.WaitForProcess(blockHash[:], shardBlocks, ProcessTypeQueue) {
		return false
	}
	if err := saver.NewBlock(p.Verbose).SaveBlock(block.Header); err != nil {
		jerr.Get("error saving block for lead node", err).Print()
		return false
	}
	if !p.WaitForProcess(blockHash[:], shardBlocks, ProcessTypeTx) {
		return false
	}
	if !p.WaitForProcess(blockHash[:], shardBlocks, ProcessTypeUtxo) {
		return false
	}
	if !p.WaitForProcess(blockHash[:], shardBlocks, ProcessTypeMeta) {
		return false
	}
	jlog.Logf("Processed block: %s %s, %d txs, size: %s\n",
		blockHash, block.Header.Timestamp.Format("2006-01-02 15:04:05"), len(block.Transactions),
		jfmt.AddCommasInt(block.SerializeSize()))
	return true
}

type ProcessType string

const (
	ProcessTypeQueue ProcessType = "queue"
	ProcessTypeTx    ProcessType = "tx"
	ProcessTypeUtxo  ProcessType = "utxo"
	ProcessTypeMeta  ProcessType = "meta"
)

func (p *Processor) WaitForProcess(blockHash []byte, shardBlocks map[uint32]*wire.MsgBlock, processType ProcessType) bool {
	var wg sync.WaitGroup
	var hadError bool
	for _, client := range p.Clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			if _, ok := shardBlocks[client.Config.Shard]; !ok {
				return
			}
			var err error
			switch processType {
			case ProcessTypeQueue:
				_, err = client.Client.Queue(context.Background(), &cluster_pb.QueueReq{
					Block: memo.GetRawBlock(*shardBlocks[client.Config.Shard]),
				})
			case ProcessTypeTx:
				_, err = client.Client.SaveTxs(context.Background(), &cluster_pb.ProcessReq{BlockHash: blockHash[:]})
			case ProcessTypeUtxo:
				_, err = client.Client.SaveUtxos(context.Background(), &cluster_pb.ProcessReq{BlockHash: blockHash[:]})
			case ProcessTypeMeta:
				_, err = client.Client.SaveMeta(context.Background(), &cluster_pb.ProcessReq{BlockHash: blockHash[:]})
			}
			if err != nil {
				hadError = true
				p.ErrorChan <- ShardError{
					Shard: client.Config.Int(),
					Error: jerr.Getf(err, "error cluster shard process: %s - %d", processType, client.Config.Shard),
				}
			}
		}(client)
	}
	wg.Wait()
	return !hadError
}

func (p *Processor) Stop() {
	if p.On {
		p.On = false
		close(p.StopChan)
		p.Node.Stop()
	}
}

func NewProcessor(verbose bool, clients map[int]*Client, errorChan chan ShardError) *Processor {
	return &Processor{
		Verbose:   verbose,
		Clients:   clients,
		ErrorChan: errorChan,
	}
}
