package server

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

const (
	DefaultPort = 26780
)

type Queue struct {
	gen.UnimplementedQueueServer
}

func (q *Queue) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", DefaultPort))
	if err != nil {
		return jerr.Get("failed to listen", err)
	}
	server := grpc.NewServer(grpc.MaxRecvMsgSize(32*10e6), grpc.MaxSendMsgSize(32*10e6))
	gen.RegisterQueueServer(server, q)
	reflection.Register(server)
	err = server.Serve(lis)
	if err != nil {
		return jerr.Get("failed to serve", err)
	}
	return jerr.New("queue server disconnected")
}

func NewQueue() *Queue {
	return &Queue{}
}
