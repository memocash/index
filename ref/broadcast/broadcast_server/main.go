package broadcast_server

import (
	"context"
	"fmt"
	"github.com/memocash/index/ref/broadcast/gen/broadcast_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server struct {
	Port             int
	BroadcastHandler func(context.Context, []byte) error
	listener         net.Listener
	grpc             *grpc.Server
	broadcast_pb.UnimplementedBroadcastServer
}

func (s *Server) BroadcastTx(ctx context.Context, request *broadcast_pb.BroadcastRequest) (*broadcast_pb.BroadcastReply, error) {
	if err := s.BroadcastHandler(ctx, request.Raw); err != nil {
		return nil, fmt.Errorf("error with broadcast tx handler; %w", err)
	}
	return new(broadcast_pb.BroadcastReply), nil
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return fmt.Errorf("error starting broadcast server; %w", err)
	}
	// Serve always returns an error
	return fmt.Errorf("error serving broadcast server; %w", s.Serve())
}

func (s *Server) Start() error {
	var err error
	if s.listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.Port)); err != nil {
		return fmt.Errorf("failed to listen broadcast; %w", err)
	}
	s.grpc = grpc.NewServer()
	broadcast_pb.RegisterBroadcastServer(s.grpc, s)
	reflection.Register(s.grpc)
	return nil
}

func (s *Server) Serve() error {
	if err := s.grpc.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve broadcast; %w", err)
	}
	return fmt.Errorf("broadcast rpc server disconnected")
}

func NewServer(port int, broadcastHandler func(context.Context, []byte) error) *Server {
	return &Server{
		Port:             port,
		BroadcastHandler: broadcastHandler,
	}
}
