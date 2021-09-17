package server

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/db/proto/queue_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

const (
	DefaultPort = 26780
)

type Queue struct {
	queue_pb.UnimplementedQueueServer
}

func (q *Queue) SaveMessages(_ context.Context, msg *queue_pb.Messages) (*queue_pb.ErrorReply, error) {
	jlog.Logf("Received %d messages\n", len(msg.Messages))
	for _, message := range msg.Messages {
		jlog.Logf("message: %x %s\n", message.Uid, message.Message)
	}
	return nil, nil
}

func (q *Queue) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", DefaultPort))
	if err != nil {
		return jerr.Get("failed to listen", err)
	}
	server := grpc.NewServer(grpc.MaxRecvMsgSize(32*10e6), grpc.MaxSendMsgSize(32*10e6))
	queue_pb.RegisterQueueServer(server, q)
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
