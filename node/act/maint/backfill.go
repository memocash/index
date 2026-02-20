package maint

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/obj/saver"
	nodePeer "github.com/memocash/index/node/peer"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/cluster/lead"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"google.golang.org/grpc"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

type Backfill struct {
	Start   int64
	End     int64
	Verbose bool
	peer    *peer.Peer
	clients map[int]*lead.Client
	ctx     context.Context
	lastReq *wire.InvVect
}

func (b *Backfill) Run() error {
	b.ctx = context.Background()
	// Look up block hash at start-1 to use as block locator
	startHeight := b.Start - 1
	heightBlock, err := chain.GetHeightBlockSingle(b.ctx, startHeight)
	if err != nil {
		return fmt.Errorf("error getting height block for start height %d (headers must be scanned first); %w",
			startHeight, err)
	}
	log.Printf("Found block at height %d: %x\n", startHeight, heightBlock.BlockHash)
	// Set up cluster shard clients
	b.clients = make(map[int]*lead.Client)
	for _, clusterShard := range config.GetClusterShards() {
		conn, err := grpc.Dial(clusterShard.GetHost(), grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("error connecting to cluster shard; %w", err)
		}
		b.clients[clusterShard.Int()] = &lead.Client{
			Config: clusterShard,
			Client: cluster_pb.NewClusterClient(conn),
		}
	}
	// Connect to BCH node
	nodePeer.SetBtcdLogLevel()
	connectionString := config.GetNodeHost()
	newPeer, err := peer.NewOutboundPeer(&peer.Config{
		UserAgentName:    "memo-index",
		UserAgentVersion: "0.3.0",
		ChainParams:      wallet.GetMainNetParams(),
		Listeners: peer.MessageListeners{
			OnVerAck:  b.OnVerAck,
			OnHeaders: b.OnHeaders,
			OnBlock:   b.OnBlock,
			OnVersion: b.OnVersion,
		},
	}, connectionString)
	if err != nil {
		return fmt.Errorf("error getting new outbound peer; %w", err)
	}
	b.peer = newPeer
	log.Printf("Starting backfill from height %s to %s, connecting to: %s\n",
		jfmt.AddCommas(b.Start), jfmt.AddCommas(b.End), connectionString)
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		return fmt.Errorf("error getting network connection; %w", err)
	}
	newPeer.AssociateConnection(conn)
	newPeer.WaitForDisconnect()
	return nil
}

func (b *Backfill) OnVerAck(_ *peer.Peer, _ *wire.MsgVerAck) {
	heightBlock, err := chain.GetHeightBlockSingle(b.ctx, b.Start-1)
	if err != nil {
		log.Fatalf("error getting height block for backfill start; %v", err)
	}
	blockHash := chainhash.Hash(heightBlock.BlockHash)
	msgGetHeaders := wire.NewMsgGetHeaders()
	msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, &blockHash)
	b.peer.QueueMessage(msgGetHeaders, nil)
}

func (b *Backfill) OnHeaders(_ *peer.Peer, msg *wire.MsgHeaders) {
	if len(msg.Headers) == 0 {
		log.Println("No more headers, backfill complete")
		b.peer.Disconnect()
		return
	}
	msgGetData := wire.NewMsgGetData()
	for _, blockHeader := range msg.Headers {
		blockHash := blockHeader.BlockHash()
		err := msgGetData.AddInvVect(&wire.InvVect{
			Type: wire.InvTypeBlock,
			Hash: blockHash,
		})
		if err != nil {
			log.Fatalf("error adding block inventory vector; %v", err)
		}
	}
	if len(msgGetData.InvList) > 0 {
		b.lastReq = msgGetData.InvList[len(msgGetData.InvList)-1]
		b.peer.QueueMessage(msgGetData, nil)
	}
}

func (b *Backfill) OnBlock(_ *peer.Peer, msg *wire.MsgBlock, _ []byte) {
	blockHash := msg.BlockHash()
	// Look up this block's height
	blockHeight, err := chain.GetBlockHeight(b.ctx, blockHash)
	if err != nil {
		log.Fatalf("error getting block height for %s; %v", blockHash, err)
	}
	height := blockHeight.Height
	if height > b.End {
		log.Printf("Reached end height %s, stopping backfill\n", jfmt.AddCommas(b.End))
		b.peer.Disconnect()
		return
	}
	seen := time.Now()
	if msg.Header.Timestamp.Before(seen) {
		seen = msg.Header.Timestamp
	}
	block := dbi.WireBlockToBlock(msg)
	// Save block metadata via block saver
	blockSaver := saver.NewBlock(b.ctx, b.Verbose)
	blockInfo := dbi.BlockInfo{
		Header:  msg.Header,
		Size:    int64(msg.SerializeSize()),
		TxCount: len(msg.Transactions),
	}
	if err := blockSaver.SaveBlock(blockInfo); err != nil {
		log.Fatalf("error saving block for backfill; %v", err)
	}
	// Partition txs by hash to shards
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
	// Send to shard clients
	var wg sync.WaitGroup
	for _, c := range b.clients {
		if _, ok := shardBlocks[c.Config.Shard]; !ok {
			continue
		}
		wg.Add(1)
		go func(c *lead.Client) {
			defer wg.Done()
			if err := lead.ExecWithRetry(func() error {
				if _, err := c.Client.SaveTxs(b.ctx, &cluster_pb.SaveReq{
					Block:     shardBlocks[c.Config.Shard],
					IsInitial: true,
					Height:    height,
					Seen:      seen.UnixNano(),
				}, grpc.MaxCallSendMsgSize(8*math.MaxInt32)); err != nil {
					return fmt.Errorf("error saving block shard txs; %w", err)
				}
				return nil
			}); err != nil {
				log.Fatalf("error saving to shard %d; %v", c.Config.Shard, err)
			}
		}(c)
	}
	wg.Wait()
	log.Printf("Backfilled block %s: %s %s, %7s txs, size: %14s\n",
		jfmt.AddCommas(height), blockHash, msg.Header.Timestamp.Format("2006-01-02 15:04:05"),
		jfmt.AddCommasInt(blockInfo.TxCount), jfmt.AddCommasInt(int(blockInfo.Size)))
	// Request next batch of headers when we've processed the last block in the batch
	if b.lastReq != nil && blockHash == b.lastReq.Hash {
		msgGetHeaders := wire.NewMsgGetHeaders()
		msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, &blockHash)
		b.peer.QueueMessage(msgGetHeaders, nil)
	}
}

func (b *Backfill) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) {
	log.Printf("Connected to peer: %s (last block: %d)\n", msg.UserAgent, msg.LastBlock)
}

func NewBackfill(start, end int64) *Backfill {
	return &Backfill{
		Start: start,
		End:   end,
	}
}
