package server

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/metric"
	"github.com/memocash/index/db/proto/queue_pb"
	"github.com/memocash/index/db/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type Server struct {
	Port        int
	Shard       uint
	Stopped     bool
	listener    net.Listener
	MsgDoneChan chan *MsgDone
	Timeout     time.Duration
	Grpc        *grpc.Server
	queue_pb.UnimplementedQueueServer
}

func (s *Server) SaveMessages(_ context.Context, messages *queue_pb.Messages) (*queue_pb.ErrorReply, error) {
	var msgs []*Msg
	for _, message := range messages.Messages {
		msgs = append(msgs, &Msg{
			Uid:     message.Uid,
			Topic:   message.Topic,
			Message: message.Message,
		})
	}
	err := s.queueSaveMessage(msgs)
	var errMsg string
	if err != nil {
		errMsg = fmt.Errorf("error queueing message server save; %w", err).Error()
	}
	return &queue_pb.ErrorReply{
		Error: errMsg,
	}, nil
}

func (s *Server) DeleteMessages(ctx context.Context, request *queue_pb.MessageUids) (*queue_pb.ErrorReply, error) {
	if err := store.DeleteMessages(request.GetTopic(), s.Shard, request.GetUids()); err != nil {
		return nil, fmt.Errorf("error deleting messages for topic; %w", err)
	}
	return &queue_pb.ErrorReply{}, nil
}

func (s *Server) StartMessageChan() {
	s.MsgDoneChan = make(chan *MsgDone)
	for {
		msgDone := <-s.MsgDoneChan
		msgDone.Done <- s.execSaveMessage(msgDone)
	}
}

func (s *Server) execSaveMessage(msgDone *MsgDone) error {
	err := s.SaveMsgs(msgDone.Msgs)
	if err != nil {
		return fmt.Errorf("error setting message; %w", err)
	}
	return nil
}

func (s *Server) SaveMsgs(msgs []*Msg) error {
	var topicMessagesToSave = make(map[string][]*store.Message)
	for _, msg := range msgs {
		topicMessagesToSave[msg.Topic] = append(topicMessagesToSave[msg.Topic], &store.Message{
			Uid:     msg.Uid,
			Message: msg.Message,
		})
	}
	for topic, messagesToSave := range topicMessagesToSave {
		err := store.SaveMessages(topic, s.Shard, messagesToSave)
		if err != nil {
			return fmt.Errorf("error saving messages for topic: %s; %w", topic, err)
		}
		for _, message := range messagesToSave {
			ReceiveNew(s.Shard, topic, message.Uid)
		}
	}
	return nil
}

func (s *Server) queueSaveMessage(msgs []*Msg) error {
	var timeout = s.Timeout
	if timeout == 0 {
		timeout = client.DefaultSetTimeout
	}
	msgDone := NewMsgDone(msgs)
	select {
	case s.MsgDoneChan <- msgDone:
		err := <-msgDone.Done
		if err != nil {
			return fmt.Errorf("error queueing message server queue; %w", err)
		}
		return nil
	case <-time.NewTimer(timeout).C:
		return fmt.Errorf("error queue message timeout (%s)", timeout)
	}
}

func (s *Server) GetMessage(_ context.Context, request *queue_pb.RequestSingle) (*queue_pb.Message, error) {
	message, err := store.GetMessage(request.Topic, s.Shard, request.Uid)
	if err != nil && !store.IsNotFoundError(err) {
		return nil, fmt.Errorf("error getting message for topic: %s, uid: %x; %w", request.Topic, request.Uid, err)
	}
	if message == nil {
		return &queue_pb.Message{}, nil
	}
	return &queue_pb.Message{
		Topic:   request.Topic,
		Uid:     message.Uid,
		Message: message.Message,
	}, nil
}

func (s *Server) GetMessages(ctx context.Context, request *queue_pb.Request) (*queue_pb.Messages, error) {
	var messages []*store.Message
	var err error
	if len(request.Uids) > 0 {
		messages, err = store.GetMessagesByUids(request.Topic, s.Shard, request.Uids)
		if err != nil {
			return nil, fmt.Errorf("error getting messages by uids; %w", err)
		}
	} else {
		for i := 0; i < 2; i++ {
			var requestByPrefixes = store.RequestByPrefixes{
				Topic:  request.Topic,
				Shard:  s.Shard,
				Max:    int(request.Max),
				Newest: request.Newest,
			}
			for _, prefix := range request.Prefixes {
				requestByPrefixes.Prefixes = append(requestByPrefixes.Prefixes, store.Prefix{
					Prefix: prefix,
					Start:  request.Start,
				})
			}
			messages, err = store.GetByPrefixes(requestByPrefixes)
			if err != nil {
				return nil, fmt.Errorf("error getting messages for topic: %s (shard %d); %w", request.Topic, s.Shard, err)
			}
			if len(messages) == 0 && request.Wait && i == 0 {
				metric.AddTopicListen(metric.TopicListen{Topic: request.Topic})
				if err := ListenSingle(ctx, s.Shard, request.Topic, request.Start, request.Prefixes); err != nil {
					return nil, fmt.Errorf("error listening for new topic item; %w", err)
				}
			} else {
				break
			}
		}
	}
	var queueMessages = make([]*queue_pb.Message, len(messages))
	for i := range messages {
		queueMessages[i] = &queue_pb.Message{
			Topic:   request.Topic,
			Uid:     messages[i].Uid,
			Message: messages[i].Message,
		}
	}
	return &queue_pb.Messages{
		Messages: queueMessages,
	}, nil
}

func (s *Server) GetStreamMessages(request *queue_pb.RequestStream, server queue_pb.Queue_GetStreamMessagesServer) error {
	ctx := server.Context()
	metric.AddTopicListen(metric.TopicListen{Topic: request.Topic})
	uidChan := Listen(ctx, s.Shard, request.Topic, request.Prefixes)
	for {
		select {
		case <-ctx.Done():
			return nil
		case uid, ok := <-uidChan:
			if !ok {
				return nil
			}
			message, err := store.GetMessage(request.Topic, s.Shard, uid)
			if err != nil {
				return fmt.Errorf("error getting stream message for topic: %s; %w", request.Topic, err)
			}
			if message == nil {
				return fmt.Errorf("error nil message from store for stream, shard: %d, topic: %s, uid: %x",
					s.Shard, request.Topic, uid)
			}
			server.Send(&queue_pb.Message{
				Uid:     uid,
				Topic:   request.Topic,
				Message: message.Message,
			})
		}
	}
}

func (s *Server) GetMessageCount(ctx context.Context, request *queue_pb.CountRequest) (*queue_pb.TopicCount, error) {
	count, err := store.GetCount(request.Topic, request.Prefix, s.Shard)
	if err != nil {
		return nil, fmt.Errorf("error getting db count for topic; %w", err)
	}
	return &queue_pb.TopicCount{
		Count: count,
	}, nil
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return fmt.Errorf("error starting db server; %w", err)
	}
	// Serve always returns an error
	return fmt.Errorf("error serving db server; %w", s.Serve())
}

func (s *Server) Start() error {
	s.Stopped = false
	var err error
	if s.listener, err = net.Listen("tcp", GetListenHost(s.Port)); err != nil {
		return fmt.Errorf("failed to listen; %w", err)
	}
	go s.StartMessageChan()
	s.Grpc = grpc.NewServer(grpc.MaxRecvMsgSize(client.MaxMessageSize), grpc.MaxSendMsgSize(client.MaxMessageSize))
	queue_pb.RegisterQueueServer(s.Grpc, s)
	reflection.Register(s.Grpc)
	return nil
}

func (s *Server) Serve() error {
	if err := s.Grpc.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve; %w", err)
	}
	return fmt.Errorf("queue server disconnected")
}

func (s *Server) Stop() {
	if s.Grpc != nil && !s.Stopped {
		s.Stopped = true
		s.Grpc.Stop()
	}
}

func NewServer(port int, shard uint) *Server {
	return &Server{
		Port:  port,
		Shard: shard,
	}
}
