package shard

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/server"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type ProcessBlock struct {
	Block *wire.MsgBlock
	Added time.Time
}

type Shard struct {
	Id       int
	Verbose  bool
	Error    chan error
	listener net.Listener
	grpc     *grpc.Server
	TxSaver  dbi.TxSave
	OutSaver dbi.TxSave
	Blocks   map[chainhash.Hash]ProcessBlock
	cluster_pb.UnimplementedClusterServer
}

func (s *Shard) CheckProcessBlocks() {
	for blockHash, block := range s.Blocks {
		if time.Since(block.Added) > time.Minute*5 {
			jlog.Logf("block not processed, removing from shard: %s\n", blockHash)
			delete(s.Blocks, blockHash)
		}
	}
	for len(s.Blocks) > 10 {
		var oldest ProcessBlock
		for _, block := range s.Blocks {
			if oldest.Added.IsZero() || block.Added.Before(oldest.Added) {
				oldest = block
			}
		}
		jlog.Logf("too many blocks in shard, removing oldest: %s\n", oldest.Block.BlockHash())
		delete(s.Blocks, oldest.Block.BlockHash())
	}
}

func (s *Shard) Run() error {
	s.Error = make(chan error)
	var err error
	clusterConfig := config.GetShardConfig(uint32(s.Id), config.GetClusterShards())
	if s.listener, err = net.Listen("tcp", clusterConfig.GetHost()); err != nil {
		return jerr.Get("failed to listen cluster shard", err)
	}
	s.grpc = grpc.NewServer()
	cluster_pb.RegisterClusterServer(s.grpc, s)
	reflection.Register(s.grpc)
	go func() {
		s.Error <- jerr.Get("failed to serve cluster shard", s.grpc.Serve(s.listener))
	}()
	queueShards := config.GetQueueShards()
	if len(queueShards) < s.Id {
		return jerr.Newf("fatal error shard specified greater than num queue shards: %d %d", s.Id, len(queueShards))
	}
	queueServer := server.NewServer(uint(queueShards[s.Id].Port), uint(s.Id))
	go func() {
		jlog.Logf("Starting queue server shard %d on port %d...\n", queueServer.Shard, queueServer.Port)
		s.Error <- jerr.Getf(queueServer.Run(), "error running queue server for shard: %d", s.Id)
	}()
	return <-s.Error
}

func (s *Shard) Ping(ctx context.Context, req *cluster_pb.PingReq) (*cluster_pb.PingResp, error) {
	jlog.Logf("received ping, nonce: %d\n", req.Nonce)
	return &cluster_pb.PingResp{
		Nonce: uint64(time.Now().UnixNano()),
	}, nil
}

func (s *Shard) Queue(ctx context.Context, req *cluster_pb.QueueReq) (*cluster_pb.EmptyResp, error) {
	block, err := memo.GetBlockFromRaw(req.Block)
	if err != nil {
		return nil, jerr.Get("error getting block from raw", err)
	}
	if s.Verbose {
		jlog.Logf("received queue shard txs, block: %s, txs: %d\n", block.BlockHash(), len(block.Transactions))
	}
	if err := s.TxSaver.SaveTxs(block); err != nil {
		return nil, jerr.Get("error saving block txs", err)
	}
	s.Blocks[block.BlockHash()] = ProcessBlock{
		Block: block,
		Added: time.Now(),
	}
	s.CheckProcessBlocks()
	if s.Verbose {
		jlog.Logf("finished queueing shard txs, block: %s\n", block.BlockHash())
	}
	return &cluster_pb.EmptyResp{}, nil
}

func (s *Shard) Process(ctx context.Context, req *cluster_pb.ProcessReq) (*cluster_pb.EmptyResp, error) {
	blockHash, err := chainhash.NewHash(req.BlockHash)
	if err != nil {
		return nil, jerr.Get("error getting block hash for shard process", err)
	}
	processBlock, ok := s.Blocks[*blockHash]
	if !ok {
		return nil, jerr.Newf("block not found for shard process: %s", blockHash.String())
	}
	delete(s.Blocks, *blockHash)
	if s.Verbose {
		jlog.Logf("received process, block: %s\n", blockHash)
	}
	if err := s.OutSaver.SaveTxs(processBlock.Block); err != nil {
		return nil, jerr.Get("error saving block txs", err)
	}
	if s.Verbose {
		jlog.Logf("finished processing, block: %s\n", processBlock.Block.BlockHash())
	}
	return &cluster_pb.EmptyResp{}, nil
}

func NewShard(shardId int, verbose bool) *Shard {
	return &Shard{
		Id:       shardId,
		Verbose:  verbose,
		TxSaver:  saver.NewCombinedTx(verbose),
		OutSaver: saver.NewCombinedOutput(verbose),
		Blocks:   make(map[chainhash.Hash]ProcessBlock),
	}
}
