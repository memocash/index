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
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Processor struct {
	Clients     map[int]*Client
	ErrorChan   chan error
	BlockNode   *Node
	MemPoolNode *Node
	Verbose     bool
	Synced      bool
}

func (p *Processor) Run() error {
	p.Clients = make(map[int]*Client)
	clusterShards := config.GetClusterShards()
	for _, clusterShard := range clusterShards {
		conn, err := grpc.Dial(clusterShard.GetHost(), grpc.WithInsecure())
		if err != nil {
			return jerr.Get("error did not connect cluster client", err)
		}
		p.Clients[clusterShard.Int()] = &Client{
			Config: clusterShard,
			Client: cluster_pb.NewClusterClient(conn)}
	}
	var syncStatusComplete *item.SyncStatus
	if err := ExecWithRetry(func() error {
		var err error
		syncStatusComplete, err = item.GetSyncStatus(item.SyncStatusComplete)
		if err != nil && !client.IsEntryNotFoundError(err) {
			return jerr.Get("error getting sync status complete", err)
		}
		return nil
	}); err != nil {
		return jerr.Get("error getting sync status complete exec with retry", err)
	}
	if syncStatusComplete != nil {
		p.Synced = true
		go func() {
			p.MemPoolNode = NewNode()
			p.MemPoolNode.Start(true, p.Synced)
			jlog.Logf("Started mempool node...\n")
			for p.ProcessBlock(<-p.MemPoolNode.NewBlock, "mempool") {
			}
			jlog.Log("Stopping mempool node")
		}()
	}
	p.BlockNode = NewNode()
	p.BlockNode.Start(false, p.Synced)
	go func() {
		jlog.Logf("Started block node...\n")
		for {
			select {
			case block := <-p.BlockNode.NewBlock:
				if p.ProcessBlock(block, "block node") {
					continue
				}
			case <-p.BlockNode.SyncDone:
				jlog.Logf("Node sync done\n")
				p.Synced = true
				recentBlock, err := chain.GetRecentHeightBlock()
				if err != nil {
					p.ErrorChan <- jerr.Get("error getting recent height block", err)
					break
				}
				if err := db.Save([]db.Object{&item.SyncStatus{
					Name:   item.SyncStatusComplete,
					Height: recentBlock.Height,
				}}); err != nil {
					p.ErrorChan <- jerr.Get("error setting sync status complete", err)
					break
				}
				if err := p.Run(); err != nil {
					p.ErrorChan <- jerr.Get("error starting lead processor after block sync complete", err)
					break
				}
			}
			jlog.Log("Stopping block node")
			return
		}
	}()
	return jerr.Get("error lead processing run", <-p.ErrorChan)
}

func (p *Processor) ProcessBlock(block *dbi.Block, loc string) bool {
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
		jlog.Logf("Saved block (%s): %s %s, %7s txs, size: %14s\n", loc,
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
			if err := ExecWithRetry(func() error {
				if _, err := c.Client.SaveTxs(context.Background(), &cluster_pb.SaveReq{
					Block:     shardBlocks[c.Config.Shard],
					IsInitial: !p.Synced,
					Height:    height,
					Seen:      seen.UnixNano(),
				}); err != nil {
					return jerr.Get("error saving block shard txs", err)
				}
				return nil
			}); err != nil {
				p.ErrorChan <- jerr.Getf(err, "error client exec with retry save txs: %d", c.Config.Shard)
			}
		}(c)
	}
	wg.Wait()
	return !hadError
}

func NewProcessor(verbose bool) *Processor {
	return &Processor{
		ErrorChan: make(chan error),
		Verbose:   verbose,
	}
}
