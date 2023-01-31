package lead

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/dbi"
	"sync"
	"time"
)

type Processor struct {
	On          bool
	StopChan    chan struct{}
	Clients     map[int]*Client
	ErrorChan   chan ShardError
	BlockNode   *Node
	MemPoolNode *Node
	Verbose     bool
	Synced      bool
}

func (p *Processor) Start() error {
	if p.On {
		return nil
	}
	p.On = true
	syncStatusTxs, err := item.GetSyncStatus(item.SyncStatusSaveTxs)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return jerr.Get("error getting sync status txs", err)
	}
	if syncStatusTxs != nil {
		p.Synced = true
		go func() {
			p.MemPoolNode = NewNode()
			p.MemPoolNode.Start(true, p.Synced)
			jlog.Logf("Started mempool node...\n")
			for p.ProcessBlock(<-p.MemPoolNode.NewBlock) {
			}
		}()
	}
	p.StopChan = make(chan struct{})
	p.BlockNode = NewNode()
	p.BlockNode.Start(false, p.Synced)
	go func() {
		jlog.Logf("Started block node...\n")
		for {
			select {
			case block := <-p.BlockNode.NewBlock:
				if p.ProcessBlock(block) {
					continue
				}
			case <-p.BlockNode.SyncDone:
				jlog.Logf("Node sync done\n")
				p.Synced = true
				recentBlock, err := chain.GetRecentHeightBlock()
				if err != nil {
					jerr.Get("error getting recent height block", err).Fatal()
				}
				if err := db.Save([]db.Object{&item.SyncStatus{
					Name:   item.SyncStatusSaveTxs,
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

func (p *Processor) ProcessBlock(block *dbi.Block) bool {
	if !p.On {
		return false
	}
	seen := time.Now()
	if block.HasHeader() && block.Header.Timestamp.Before(seen) {
		seen = block.Header.Timestamp
	}
	var shardBlocks = make(map[uint32]*cluster_pb.Block)
	for i, tx := range block.Transactions {
		shard := db.GetShardByte32(tx.Hash[:])
		if _, ok := shardBlocks[shard]; !ok {
			shardBlocks[shard] = &cluster_pb.Block{
				Header: memo.GetRawBlockHeader(block.Header),
			}
		}
		shardBlocks[shard].Txs = append(shardBlocks[shard].Txs, &cluster_pb.Tx{
			Index: uint32(i),
			Raw:   memo.GetRaw(tx.MsgTx),
		})
	}
	blockHash := block.Header.BlockHash()
	blockInfo := dbi.BlockInfo{
		Header:  block.Header,
		Size:    block.Size(),
		TxCount: len(block.Transactions),
	}
	var height int64
	if dbi.BlockHeaderSet(block.Header) {
		blockSaver := saver.NewBlock(p.Verbose)
		if err := blockSaver.SaveBlock(blockInfo); err != nil {
			jerr.Get("error saving block for lead node", err).Print()
			return false
		}
		if blockSaver.NewHeight == 0 {
			// A block without a height can happen if you receive a new block while syncing, ignore it, don't save TXs.
			return true
		}
		height = blockSaver.NewHeight
	}
	if !p.SaveBlockShards(height, seen, shardBlocks) {
		return false
	}
	if dbi.BlockHeaderSet(block.Header) {
		jlog.Logf("Saved block: %s %s, %7s txs, size: %14s\n",
			blockHash, block.Header.Timestamp.Format("2006-01-02 15:04:05"), jfmt.AddCommasInt(blockInfo.TxCount),
			jfmt.AddCommasInt(int(blockInfo.Size)))
	}
	return true
}

func (p *Processor) SaveBlockShards(height int64, seen time.Time, shardBlocks map[uint32]*cluster_pb.Block) bool {
	var wg sync.WaitGroup
	var hadError bool
	for _, c := range p.Clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			if _, ok := shardBlocks[c.Config.Shard]; !ok {
				return
			}
			if _, err := c.Client.SaveTxs(context.Background(), &cluster_pb.SaveReq{
				Block:     shardBlocks[c.Config.Shard],
				IsInitial: !p.Synced,
				Height:    height,
				Seen:      seen.UnixNano(),
			}); err != nil {
				hadError = true
				p.ErrorChan <- ShardError{
					Shard: c.Config.Int(),
					Error: jerr.Getf(err, "error cluster shard process: %d", c.Config.Shard),
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
		p.BlockNode.Stop()
		if p.MemPoolNode != nil {
			p.MemPoolNode.Stop()
		}
	}
}

func NewProcessor(verbose bool, clients map[int]*Client, errorChan chan ShardError) *Processor {
	return &Processor{
		Verbose:   verbose,
		Clients:   clients,
		ErrorChan: errorChan,
	}
}
