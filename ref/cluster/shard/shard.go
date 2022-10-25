package shard

import (
	"context"
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

type Shard struct {
	Id       int
	Error    chan error
	listener net.Listener
	grpc     *grpc.Server
	TxSaver  dbi.TxSave
	cluster_pb.UnimplementedClusterServer
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

func (s *Shard) Process(ctx context.Context, req *cluster_pb.ProcessReq) (*cluster_pb.ProcessResp, error) {
	block, err := memo.GetBlockFromRaw(req.Block)
	if err != nil {
		return nil, jerr.Get("error getting block from raw", err)
	}
	jlog.Logf("received process, block: %s, txs: %d\n", block.BlockHash(), len(block.Transactions))
	if err := s.TxSaver.SaveTxs(dbi.GetBlockWithHeight(block, req.Height)); err != nil {
		return nil, jerr.Get("error saving block txs", err)
	}
	jlog.Logf("finished processing, block: %s\n", block.BlockHash())
	return &cluster_pb.ProcessResp{}, nil
}

func NewShard(shardId int) *Shard {
	return &Shard{
		Id:      shardId,
		TxSaver: saver.NewCombinedAll(false),
	}
}
