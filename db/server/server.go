package server

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
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
		errMsg = jerr.Get("error queueing message", err).Error()
	}
	return &queue_pb.ErrorReply{
		Error: errMsg,
	}, nil
}

func (s *Server) DeleteMessages(ctx context.Context, request *queue_pb.MessageUids) (*queue_pb.ErrorReply, error) {
	if err := store.DeleteMessages(request.GetTopic(), s.Shard, request.GetUids()); err != nil {
		return nil, jerr.Get("error deleting messages for topic", err)
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
		return jerr.Get("error setting message", err)
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
			return jerr.Getf(err, "error saving messages for topic: %s", topic)
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
		timeout = client.DefaultGetTimeout
	}
	msgDone := NewMsgDone(msgs)
	select {
	case s.MsgDoneChan <- msgDone:
		err := <-msgDone.Done
		if err != nil {
			return jerr.Get("error queueing message", err)
		}
		return nil
	case <-time.NewTimer(timeout).C:
		return jerr.Newf("error queue message timeout (%s)", timeout)
	}
}

func (s *Server) GetMessage(_ context.Context, request *queue_pb.RequestSingle) (*queue_pb.Message, error) {
	message, err := store.GetMessage(request.Topic, s.Shard, request.Uid)
	if err != nil && !store.IsNotFoundError(err) {
		return nil, jerr.Getf(err, "error getting message for topic: %s, uid: %x", request.Topic, request.Uid)
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
			return nil, jerr.Get("error getting messages by uids", err)
		}
	} else {
		for i := 0; i < 2; i++ {
			messages, err = store.GetMessages(request.Topic, s.Shard, request.Prefixes, request.Start, int(request.Max),
				request.Newest)
			if err != nil {
				return nil, jerr.Getf(err, "error getting messages for topic: %s (shard %d)", request.Topic, s.Shard)
			}
			if len(messages) == 0 && request.Wait && i == 0 {
				if err := ListenSingle(ctx, s.Shard, request.Topic, request.Start, request.Prefixes); err != nil {
					return nil, jerr.Get("error listening for new topic item", err)
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
	uidChan := Listen(server.Context(), s.Shard, request.Topic, request.Prefixes)
	for {
		uid := <-uidChan
		if uid == nil {
			// End of stream
			return nil
		}
		message, err := store.GetMessage(request.Topic, s.Shard, uid)
		if err != nil {
			return jerr.Getf(err, "error getting stream message for topic: %s", request.Topic)
		}
		if message == nil {
			return jerr.Newf("error nil message from store for stream, shard: %d, topic: %s, uid: %x",
				s.Shard, request.Topic, uid)
		}
		server.Send(&queue_pb.Message{
			Uid:     uid,
			Topic:   request.Topic,
			Message: message.Message,
		})
	}
}

func (s *Server) GetMessageCount(ctx context.Context, request *queue_pb.CountRequest) (*queue_pb.TopicCount, error) {
	count, err := store.GetCount(request.Topic, request.Prefix, s.Shard)
	if err != nil {
		return nil, jerr.Get("error getting db count for topic", err)
	}
	return &queue_pb.TopicCount{
		Count: count,
	}, nil
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return jerr.Get("error starting db server", err)
	}
	// Serve always returns an error
	return jerr.Get("error serving db server", s.Serve())
}

func (s *Server) Start() error {
	s.Stopped = false
	var err error
	if s.listener, err = net.Listen("tcp", GetListenHost(s.Port)); err != nil {
		return jerr.Get("failed to listen", err)
	}
	go s.StartMessageChan()
	s.Grpc = grpc.NewServer(grpc.MaxRecvMsgSize(client.MaxMessageSize), grpc.MaxSendMsgSize(client.MaxMessageSize))
	queue_pb.RegisterQueueServer(s.Grpc, s)
	reflection.Register(s.Grpc)
	return nil
}

func (s *Server) Serve() error {
	if err := s.Grpc.Serve(s.listener); err != nil {
		return jerr.Get("failed to serve", err)
	}
	return jerr.New("queue server disconnected")
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
