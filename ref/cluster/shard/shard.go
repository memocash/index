package shard

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
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
	if err := s.grpc.Serve(s.listener); err != nil {
		return jerr.Get("failed to serve broadcast", err)
	}
	return <-s.Error
}

func (s *Shard) Ping(ctx context.Context, req *cluster_pb.PingReq) (*cluster_pb.PingResp, error) {
	jlog.Logf("received ping, nonce: %d\n", req.Nonce)
	return &cluster_pb.PingResp{
		Nonce: uint64(time.Now().UnixNano()),
	}, nil
}

func (s *Shard) Process(ctx context.Context, req *cluster_pb.ProcessReq) (*cluster_pb.ProcessResp, error) {
	jlog.Logf("received process, block: %x\n", req.Block)
	time.Sleep(time.Second * 5)
	jlog.Logf("finished processing, block: %x\n", req.Block)
	return &cluster_pb.ProcessResp{}, nil
}

func NewShard(shardId int) *Shard {
	return &Shard{
		Id: shardId,
	}
}
