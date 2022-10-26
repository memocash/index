package lead

import (
	"context"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"google.golang.org/grpc"
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

func (p *Processor) Start() error {
	if p.On {
		return nil
	}
	p.On = true
	syncStatusTxs, err := item.GetSyncStatus(item.SyncStatusTxs)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return jerr.Get("error getting sync status txs", err)
	}
	syncStatusUtxos, err := item.GetSyncStatus(item.SyncStatusUtxos)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return jerr.Get("error getting sync status utxos", err)
	}
	if syncStatusTxs != nil {
		go func() {
			var height int64
			if syncStatusUtxos != nil {
				height = syncStatusUtxos.Height
			} else {
				oldestHeightBlock, err := item.GetOldestHeightBlock()
				if err != nil {
					jerr.Get("error getting oldest height block", err).Fatal()
					return
				}
				height = oldestHeightBlock.Height
			}
			jlog.Logf("Starting utxo processing at height: %d\n", height)
			for {
				if height >= syncStatusTxs.Height {
					jlog.Logf("UTXO processing complete at height: %d\n", height)
					break
				}
				heightBlock, err := item.GetHeightBlockSingle(height)
				if err != nil {
					jerr.Get("error getting height block", err).Fatal()
					return
				}
				if !p.WaitForProcess(heightBlock.BlockHash, nil, ProcessTypeUtxo) {
					return
				}
				height++
			}
		}()
		return nil
	}
	p.StopChan = make(chan struct{})
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
			case <-p.Node.SyncDone:
				jlog.Logf("Node sync done\n")
				recentBlock, err := item.GetRecentHeightBlock()
				if err != nil {
					jerr.Get("error getting recent height block", err).Fatal()
				}
				if err := db.Save([]db.Object{&item.SyncStatus{
					Name:   item.SyncStatusTxs,
					Height: recentBlock.Height,
				}}); err != nil {
					jerr.Get("error setting sync status txs", err).Fatal()
				}
				p.On = false
				if err := p.Start(); err != nil {
					jerr.Get("error starting lead processor after block sync complete", err).Fatal()
				}
			case <-p.StopChan:
			}
			jlog.Log("Stopping node listener")
			return
		}
	}()
	return nil
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
	if err := saver.NewBlock(p.Verbose).SaveBlock(block.Header); err != nil {
		jerr.Get("error saving block for lead node", err).Print()
		return false
	}
	if !p.WaitForProcess(blockHash[:], shardBlocks, ProcessTypeTx) {
		return false
	}
	/*if !p.WaitForProcess(blockHash[:], shardBlocks, ProcessTypeUtxo) {
		return false
	}
	if !p.WaitForProcess(blockHash[:], shardBlocks, ProcessTypeMeta) {
		return false
	}*/
	jlog.Logf("Saved block: %s %s, %7s txs, size: %14s\n",
		blockHash, block.Header.Timestamp.Format("2006-01-02 15:04:05"), jfmt.AddCommasInt(len(block.Transactions)),
		jfmt.AddCommasInt(block.SerializeSize()))
	return true
}

type ProcessType string

const (
	ProcessTypeTx   ProcessType = "tx"
	ProcessTypeUtxo ProcessType = "utxo"
	ProcessTypeMeta ProcessType = "meta"
)

func (p *Processor) WaitForProcess(blockHash []byte, shardBlocks map[uint32]*wire.MsgBlock, processType ProcessType) bool {
	var wg sync.WaitGroup
	var hadError bool
	for _, c := range p.Clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			if _, ok := shardBlocks[c.Config.Shard]; !ok && processType == ProcessTypeTx {
				return
			}
			var err error
			switch processType {
			case ProcessTypeTx:
				_, err = c.Client.SaveTxs(context.Background(), &cluster_pb.SaveReq{
					Block: memo.GetRawBlock(*shardBlocks[c.Config.Shard]),
				}, grpc.MaxCallSendMsgSize(client.MaxMessageSize))
			case ProcessTypeUtxo:
				_, err = c.Client.SaveUtxos(context.Background(), &cluster_pb.ProcessReq{BlockHash: blockHash[:]})
			case ProcessTypeMeta:
				_, err = c.Client.SaveMeta(context.Background(), &cluster_pb.ProcessReq{BlockHash: blockHash[:]})
			}
			if err != nil {
				hadError = true
				p.ErrorChan <- ShardError{
					Shard: c.Config.Int(),
					Error: jerr.Getf(err, "error cluster shard process: %s - %d", processType, c.Config.Shard),
				}
			}
		}(c)
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
