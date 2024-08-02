package client

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/proto/queue_pb"
	"google.golang.org/grpc"
	"log"
	"time"
)

var connHandler *ConnHandler

type Client struct {
	host     string
	conn     *grpc.ClientConn
	Messages []Message
	Topics   []Topic
}

func (s *Client) SetConn() error {
	if connHandler == nil {
		connHandler = new(ConnHandler)
		connHandler.Start()
		startStats()
	}
	conn := connHandler.Get(s.host)
	if conn != nil {
		s.conn = conn
		return nil
	}
	newConn, err := grpc.Dial(s.host, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("error broadcast rpc did not connect; %w", err)
	}
	connHandler.Add(s.host, newConn)
	s.conn = newConn
	return nil
}

func (s *Client) Save(messages []*Message, timestamp time.Time) error {
	if err := s.SetConn(); err != nil {
		return fmt.Errorf("error setting connection; %w", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultSetTimeout)
	defer cancel()
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	var queueMessages = make([]*queue_pb.Message, len(messages))
	for i := 0; len(messages) > 0; i++ {
		queueMessages[i], messages = &queue_pb.Message{
			Uid:       messages[0].Uid,
			Topic:     messages[0].Topic,
			Message:   messages[0].Message,
			Timestamp: timestamp.Unix(),
		}, messages[1:]
	}
	for len(queueMessages) > 0 {
		max := len(queueMessages)
		if max > ExLargeLimit {
			max = ExLargeLimit
		}
		var queueMessagesToUse []*queue_pb.Message
		queueMessagesToUse, queueMessages = queueMessages[:max], queueMessages[max:]
		reply, err := c.SaveMessages(ctx, &queue_pb.Messages{
			Messages: queueMessagesToUse,
		}, grpc.MaxCallRecvMsgSize(MaxMessageSize), grpc.MaxCallSendMsgSize(MaxMessageSize))
		if err != nil {
			return fmt.Errorf("error saving messages and getting reply rpc: %d; %w", len(queueMessagesToUse), err)
		}
		if reply.Error != "" {
			return fmt.Errorf("error queueing message client save; %w", fmt.Errorf("%s", reply.Error))
		}
	}
	return nil
}

func (s *Client) GetByPrefixes(ctx context.Context, topic string, prefixes []Prefix, opts ...Option) error {
	c := queue_pb.NewQueueClient(s.conn)
	ctxNew, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	s.Messages = nil
	var reqPrefixes = make([]*queue_pb.RequestPrefix, len(prefixes))
	for i := range prefixes {
		reqPrefixes[i] = &queue_pb.RequestPrefix{
			Prefix: prefixes[i].Prefix,
			Start:  prefixes[i].Start,
			Limit:  prefixes[i].Limit,
		}
	}
	req := &queue_pb.RequestPrefixes{
		Topic:    topic,
		Prefixes: reqPrefixes,
	}
	for _, opt := range opts {
		opt.Apply(req)
	}
	message, err := c.GetByPrefixes(ctxNew, req, grpc.MaxCallRecvMsgSize(MaxMessageSize))
	if err != nil {
		return fmt.Errorf("error getting messages rpc; %w", err)
	}
	var messages = make([]Message, len(message.Messages))
	for i := range message.Messages {
		messages[i] = Message{
			Topic:   message.Messages[i].Topic,
			Uid:     message.Messages[i].Uid,
			Message: message.Messages[i].Message,
		}
	}
	s.Messages = append(s.Messages, messages...)
	return nil
}

func (s *Client) GetByPrefix(ctx context.Context, topic string, prefix Prefix, opts ...Option) error {
	if err := s.GetByPrefixes(ctx, topic, []Prefix{prefix}, opts...); err != nil {
		return fmt.Errorf("error getting with single prefix; %w", err)
	}
	return nil
}

func (s *Client) GetFirst(ctx context.Context, topic string) error {
	if err := s.GetByPrefix(ctx, topic, Prefix{Limit: 1}); err != nil {
		return fmt.Errorf("error getting first; %w", err)
	}
	return nil
}

func (s *Client) GetLast(ctx context.Context, topic string) error {
	if err := s.GetByPrefix(ctx, topic, Prefix{Limit: 1}, NewOptionOrder(true)); err != nil {
		return fmt.Errorf("error getting last; %w", err)
	}
	return nil
}

func (s *Client) GetAll(ctx context.Context, topic string, start []byte, opts ...Option) error {
	if err := s.GetByPrefix(ctx, topic, Prefix{Start: start}, opts...); err != nil {
		return fmt.Errorf("error getting all; %w", err)
	}
	return nil
}

func (s *Client) GetSingle(ctx context.Context, topic string, uid []byte) error {
	if err := s.SetConn(); err != nil {
		return fmt.Errorf("error setting connection; %w", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(ctx, DefaultGetTimeout)
	defer cancel()
	message, err := c.GetMessage(ctx, &queue_pb.RequestSingle{
		Topic: topic,
		Uid:   uid,
	})
	if err != nil {
		return fmt.Errorf("error getting single message rpc; %w", err)
	}
	if len(message.Uid) == 0 {
		return fmt.Errorf("empty message returned, uid empty: %d (%s); %w",
			len(message.Message), message.Topic, MessageNotSetError)
	}
	s.Messages = []Message{{
		Topic:   message.Topic,
		Uid:     message.Uid,
		Message: message.Message,
	}}
	return nil
}

func (s *Client) GetNext(ctx context.Context, topic string, start []byte) error {
	startPlusOne := jutil.CombineBytes(start, []byte{0x0})
	if err := s.GetByPrefix(ctx, topic, Prefix{
		Start: startPlusOne,
		Limit: 1,
	}); err != nil {
		return fmt.Errorf("error getting next with prefix; %w", err)
	}
	return nil
}

func (s *Client) GetSpecific(ctx context.Context, topic string, uids [][]byte) error {
	if err := s.SetConn(); err != nil {
		return fmt.Errorf("error setting connection; %w", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(ctx, DefaultGetTimeout)
	defer cancel()
	messages, err := c.GetMessages(ctx, &queue_pb.Request{
		Topic: topic,
		Uids:  uids,
	})
	if err != nil {
		return fmt.Errorf("error getting specific message rpc; %w", err)
	} else if messages == nil {
		return fmt.Errorf("error getting specific message rpc; %w", MessageNotSetError)
	}
	for _, message := range messages.Messages {
		s.Messages = append(s.Messages, Message{
			Topic:   message.Topic,
			Uid:     message.Uid,
			Message: message.Message,
		})
	}
	return nil
}

type Opts struct {
	Topic    string
	Start    []byte
	Prefixes [][]byte
	Max      uint32
	Uids     [][]byte
	Newest   bool
	Context  context.Context
}

func (s *Client) Listen(ctx context.Context, topic string, prefixes [][]byte) (chan *Message, error) {
	if err := s.SetConn(); err != nil {
		return nil, fmt.Errorf("error setting connection; %w", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(ctx, DefaultStreamTimeout)
	var request = &queue_pb.RequestStream{
		Topic:    topic,
		Prefixes: prefixes,
	}
	stream, err := c.GetStreamMessages(ctx, request)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error getting stream messages; %w", err)
	}
	var messageChan = make(chan *Message)
	go func() {
		stat := newStat()
		defer removeStat(stat)
		for {
			msg, err := stream.Recv()
			if err != nil {
				if !jerr.HasErrorPart(err, context.Canceled.Error()) {
					log.Printf("error receiving stream message; %v", err)
				}
				close(messageChan)
				cancel()
				return
			}
			messageChan <- &Message{
				Topic:   msg.Topic,
				Uid:     msg.Uid,
				Message: msg.Message,
			}
			stat.incr()
		}
	}()
	return messageChan, nil
}

func (s *Client) GetTopicCount(topic string, prefix []byte) (uint64, error) {
	if err := s.SetConn(); err != nil {
		return 0, fmt.Errorf("error setting connection; %w", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultGetTimeout)
	defer cancel()
	topicCount, err := c.GetMessageCount(ctx, &queue_pb.CountRequest{
		Topic:  topic,
		Prefix: prefix,
	})
	if err != nil {
		return 0, fmt.Errorf("error getting topic count; %w", err)
	}
	return topicCount.GetCount(), nil
}

func (s *Client) DeleteMessages(topic string, uids [][]byte) error {
	if err := s.SetConn(); err != nil {
		return fmt.Errorf("error setting connection; %w", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultSetTimeout)
	defer cancel()
	if _, err := c.DeleteMessages(ctx, &queue_pb.MessageUids{
		Topic: topic,
		Uids:  uids,
	}); err != nil {
		return fmt.Errorf("error deleting items for topics; %w", err)
	}
	return nil
}

func NewClient(host string) *Client {
	return &Client{
		host: host,
	}
}
