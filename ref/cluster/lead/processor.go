package lead

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/jchavannes/jgo/jfmt"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"google.golang.org/grpc"
)

type Processor struct {
	Context     context.Context
	Clients     map[int]*Client
	ErrorChan   chan error
	BlockNode   *BlockNode
	MempoolNode *MempoolNode
	Verbose     bool
	Synced      bool
}

func (p *Processor) Run() error {
	if err := NewScanHeaders().Run(); err != nil {
		return fmt.Errorf("error scanning block headers; %w", err)
	}
	p.Clients = make(map[int]*Client)
	clusterShards := config.GetClusterShards()
	for _, clusterShard := range clusterShards {
		conn, err := grpc.Dial(clusterShard.GetHost(), grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("error did not connect cluster client; %w", err)
		}
		p.Clients[clusterShard.Int()] = &Client{
			Config: clusterShard,
			Client: cluster_pb.NewClusterClient(conn)}
	}
	p.BlockNode = NewBlockNode()
	p.BlockNode.Start()
	for {
		select {
		case block := <-p.BlockNode.NewBlock:
			var loc string
			if p.Synced {
				loc = "block node"
			} else {
				loc = "block sync"
			}
			if !p.ProcessBlock(block, loc) {
				return fmt.Errorf("error processing block during sync")
			}
		case <-p.BlockNode.SyncDone:
			p.Synced = true
			p.MempoolNode = NewMempoolNode()
			p.MempoolNode.Start()
			go func() {
				for {
					block := <-p.MempoolNode.NewBlock
					if !p.ProcessBlock(block, "mempool") {
						p.ErrorChan <- fmt.Errorf("error processing mempool block")
						return
					}
				}
			}()
		case err := <-p.ErrorChan:
			return fmt.Errorf("error lead processing run; %w", err)
		}
	}
}

func (p *Processor) ProcessBlock(block *dbi.Block, loc string) bool {
	seen := time.Now()
	if block.HasHeader() && block.Header.Timestamp.Before(seen) {
		seen = block.Header.Timestamp
	}
	var shardBlocks = make(map[uint32]*cluster_pb.Block)
	for i, tx := range block.Transactions {
		shard := db.GetShardIdFromByte32(tx.Hash[:])
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
		blockSaver := saver.NewBlock(p.Context, p.Verbose)
		if err := blockSaver.SaveBlock(blockInfo); err != nil {
			log.Printf("error saving block for lead node; %v", err)
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
	if height > 0 {
		if err := db.Save([]db.Object{&item.SyncStatus{
			Name:   item.SyncStatusBlockHeight,
			Height: height,
		}}); err != nil {
			log.Printf("error saving sync status block height; %v", err)
			return false
		}
	}
	if dbi.BlockHeaderSet(block.Header) {
		log.Printf("Saved block (%s): %s %s, %7s txs, size: %14s\n", loc,
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
				if _, err := c.Client.SaveTxs(p.Context, &cluster_pb.SaveReq{
					Block:     shardBlocks[c.Config.Shard],
					IsInitial: !p.Synced, // Used to optimize some queries
					Height:    height,
					Seen:      seen.UnixNano(),
				}, grpc.MaxCallSendMsgSize(8*math.MaxInt32)); err != nil {
					return fmt.Errorf("error saving block shard txs; %w", err)
				}
				return nil
			}); err != nil {
				hadError = true
				p.ErrorChan <- fmt.Errorf("error client exec with retry save txs: %d; %w", c.Config.Shard, err)
			}
		}(c)
	}
	wg.Wait()
	return !hadError
}

func NewProcessor(ctx context.Context, verbose bool) *Processor {
	return &Processor{
		Context:   ctx,
		ErrorChan: make(chan error),
		Verbose:   verbose,
	}
}
