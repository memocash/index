package queue

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
	DefaultShard0Port = 26780
	DefaultShard1Port = 26781
)

type Server struct {
	Port uint
	Grpc *grpc.Server
	queue_pb.UnimplementedQueueServer
}

func (s *Server) SaveMessages(_ context.Context, msg *queue_pb.Messages) (*queue_pb.ErrorReply, error) {
	jlog.Logf("Received %d messages\n", len(msg.Messages))
	for _, message := range msg.Messages {
		jlog.Logf("message: %x %s\n", message.Uid, message.Message)
	}
	return nil, nil
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.Port))
	if err != nil {
		return jerr.Get("failed to listen", err)
	}
	s.Grpc = grpc.NewServer(grpc.MaxRecvMsgSize(32*10e6), grpc.MaxSendMsgSize(32*10e6))
	queue_pb.RegisterQueueServer(s.Grpc, s)
	reflection.Register(s.Grpc)
	if err = s.Grpc.Serve(lis); err != nil {
		return jerr.Get("failed to serve", err)
	}
	return jerr.New("queue server disconnected")
}

func (s *Server) Stop() {
	if s.Grpc != nil {
		s.Grpc.Stop()
	}
}

func NewServer(port uint) *Server {
	return &Server{
		Port: port,
	}
}
