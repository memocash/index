package client

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/proto/queue_pb"
	"google.golang.org/grpc"
	"time"
)

var connHandler *ConnHandler

type Client struct {
	host     string
	conn     *grpc.ClientConn
	Messages []Message
	Topics   []Topic
}

func (s *Client) GetLast() string {
	if len(s.Messages) == 0 {
		return ""
	}
	return hex.EncodeToString(s.Messages[len(s.Messages)-1].Uid)
}

func (s *Client) SetConn() error {
	if connHandler == nil {
		connHandler = new(ConnHandler)
		connHandler.Start()
	}
	conn := connHandler.Get(s.host)
	if conn != nil {
		s.conn = conn
		return nil
	}
	newConn, err := grpc.Dial(s.host, grpc.WithInsecure())
	if err != nil {
		return jerr.Get("error broadcast rpc did not connect", err)
	}
	connHandler.Add(s.host, newConn)
	s.conn = newConn
	return nil
}

func (s *Client) SaveSingle(message *Message, timestamp time.Time) error {
	if err := s.Save([]*Message{message}, timestamp); err != nil {
		return jerr.Get("error saving single client message", err)
	}
	return nil
}

func (s *Client) Save(messages []*Message, timestamp time.Time) error {
	if err := s.SetConn(); err != nil {
		return jerr.Get("error setting connection", err)
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
			return jerr.Getf(err, "error saving messages and getting reply rpc: %d", len(queueMessagesToUse))
		}
		if reply.Error != "" {
			return jerr.Get("error queueing message", jerr.New(reply.Error))
		}
	}
	return nil
}

func (s *Client) Get(topic string, start []byte, wait bool) error {
	if err := s.GetWOpts(Opts{
		Topic: topic,
		Start: start,
		Wait:  wait,
	}); err != nil {
		return jerr.Get("error getting with opts", err)
	}
	return nil
}

func (s *Client) GetSpecific(topic string, uids [][]byte) error {
	if err := s.GetWOpts(Opts{
		Topic: topic,
		Uids:  uids,
	}); err != nil {
		return jerr.Get("error getting with opts specific", err)
	}
	return nil
}

func (s *Client) GetByPrefixes(topic string, prefixes [][]byte) error {
	if err := s.GetWOpts(Opts{
		Topic:    topic,
		Prefixes: prefixes,
	}); err != nil {
		return jerr.Get("error getting with opts prefixes", err)
	}
	return nil
}

func (s *Client) GetByPrefix(topic string, prefix []byte) error {
	if err := s.GetWOpts(Opts{
		Topic:    topic,
		Prefixes: [][]byte{prefix},
	}); err != nil {
		return jerr.Get("error getting with opts prefix", err)
	}
	return nil
}

func (s *Client) GetSingle(topic string, uid []byte) error {
	if err := s.GetSingleContext(context.Background(), topic, uid); err != nil {
		return jerr.Getf(err, "error getting single for topic / uid: %s, %x", topic, uid)
	}
	return nil
}

func (s *Client) GetSingleContext(ctx context.Context, topic string, uid []byte) error {
	if err := s.SetConn(); err != nil {
		return jerr.Get("error setting connection", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(ctx, DefaultGetTimeout)
	defer cancel()
	message, err := c.GetMessage(ctx, &queue_pb.RequestSingle{
		Topic: topic,
		Uid:   uid,
	})
	if err != nil {
		return jerr.Get("error getting single message rpc", err)
	}
	if len(message.Uid) == 0 {
		return jerr.Getf(MessageNotSetError, "empty message returned, uid empty: %d (%s)",
			len(message.Message), message.Topic)
	}
	s.Messages = []Message{{
		Topic:   message.Topic,
		Uid:     message.Uid,
		Message: message.Message,
	}}
	return nil
}

func (s *Client) GetLarge(topic string, start []byte, wait bool, newest bool) error {
	if err := s.GetWOpts(Opts{
		Topic:  topic,
		Start:  start,
		Wait:   wait,
		Max:    LargeLimit,
		Newest: newest,
	}); err != nil {
		return jerr.Get("error getting with opts", err)
	}
	return nil
}

func (s *Client) GetNext(topic string, start []byte, wait bool, newest bool) error {
	startPlusOne := jutil.CombineBytes(start, []byte{0x0})
	if err := s.GetWOpts(Opts{
		Topic:  topic,
		Start:  startPlusOne,
		Wait:   wait,
		Max:    1,
		Newest: newest,
	}); err != nil {
		return jerr.Get("error getting next with opts", err)
	}
	return nil
}

type Opts struct {
	Topic    string
	Start    []byte
	Prefixes [][]byte
	Max      uint32
	Uids     [][]byte
	Wait     bool
	Newest   bool
	Context  context.Context
	Timeout  time.Duration
}

func (s *Client) GetWOpts(opts Opts) error {
	var optGroups []Opts
	if len(opts.Prefixes) > ExLargeLimit {
		for i := 0; i < len(opts.Prefixes); i += ExLargeLimit {
			end := i + ExLargeLimit
			if end > len(opts.Prefixes) {
				end = len(opts.Prefixes)
			}
			optGroups = append(optGroups, Opts{
				Topic:    opts.Topic,
				Start:    opts.Start,
				Prefixes: opts.Prefixes[i:end],
				Max:      opts.Max,
				Uids:     opts.Uids,
				Wait:     opts.Wait,
				Newest:   opts.Newest,
			})
		}
	} else if len(opts.Uids) > ExLargeLimit {
		for i := 0; i < len(opts.Uids); i += ExLargeLimit {
			end := i + ExLargeLimit
			if end > len(opts.Uids) {
				end = len(opts.Uids)
			}
			optGroups = append(optGroups, Opts{
				Topic:    opts.Topic,
				Prefixes: opts.Prefixes,
				Start:    opts.Start,
				Max:      opts.Max,
				Uids:     opts.Uids[i:end],
				Wait:     opts.Wait,
				Newest:   opts.Newest,
			})
		}
	} else {
		optGroups = []Opts{opts}
	}
	if err := s.SetConn(); err != nil {
		return jerr.Get("error setting connection", err)
	}
	var timeout time.Duration
	if opts.Timeout > 0 {
		timeout = opts.Timeout
	} else if !opts.Wait {
		timeout = DefaultGetTimeout
	} else {
		timeout = DefaultWaitTimeout
	}
	c := queue_pb.NewQueueClient(s.conn)
	var bgCtx = opts.Context
	if jutil.IsNil(bgCtx) {
		bgCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(bgCtx, timeout)
	defer cancel()
	s.Messages = nil
	for _, optGroup := range optGroups {
		message, err := c.GetMessages(ctx, &queue_pb.Request{
			Topic:    optGroup.Topic,
			Prefixes: optGroup.Prefixes,
			Start:    optGroup.Start,
			Max:      optGroup.Max,
			Uids:     optGroup.Uids,
			Wait:     optGroup.Wait,
			Newest:   optGroup.Newest,
		}, grpc.MaxCallRecvMsgSize(MaxMessageSize))
		if err != nil {
			return jerr.Get("error getting messages rpc", err)
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
	}
	return nil
}

func (s *Client) GetTopicList() error {
	if err := s.SetConn(); err != nil {
		return jerr.Get("error setting connection", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultSetTimeout)
	defer cancel()
	topicList, err := c.GetTopicList(ctx, new(queue_pb.EmptyRequest))
	if err != nil {
		return jerr.Get("error getting topic list", err)
	}
	var topics = make([]Topic, len(topicList.Topics))
	for i := range topicList.Topics {
		topics[i] = Topic{
			Name: topicList.Topics[i].Name,
			Size: topicList.Topics[i].Count,
		}
	}
	s.Topics = topics
	return nil
}

func (s *Client) Listen(ctx context.Context, topic string, prefixes [][]byte) (chan *Message, error) {
	messageChan, err := s.ListenOpts(Opts{
		Context:  ctx,
		Topic:    topic,
		Prefixes: prefixes,
	})
	if err != nil {
		return nil, jerr.Get("error getting message chan with opts", err)
	}
	return messageChan, nil
}

func (s *Client) ListenOpts(opts Opts) (chan *Message, error) {
	if err := s.SetConn(); err != nil {
		return nil, jerr.Get("error setting connection", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(opts.Context, DefaultStreamTimeout)
	var request = &queue_pb.RequestStream{
		Topic:    opts.Topic,
		Prefixes: opts.Prefixes,
	}
	stream, err := c.GetStreamMessages(ctx, request)
	if err != nil {
		cancel()
		return nil, jerr.Get("error getting stream messages", err)
	}
	var messageChan = make(chan *Message)
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				if !jerr.HasErrorPart(err, context.Canceled.Error()) {
					jerr.Get("error receiving stream message", err).Print()
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
		}
	}()
	return messageChan, nil
}

func (s *Client) GetTopicCount(topic string, prefix []byte) (uint64, error) {
	if err := s.SetConn(); err != nil {
		return 0, jerr.Get("error setting connection", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultSetTimeout)
	defer cancel()
	topicCount, err := c.GetMessageCount(ctx, &queue_pb.CountRequest{
		Topic:  topic,
		Prefix: prefix,
	})
	if err != nil {
		return 0, jerr.Get("error getting topic count", err)
	}
	return topicCount.GetCount(), nil
}

func (s *Client) DeleteMessages(topic string, uids [][]byte) error {
	if err := s.SetConn(); err != nil {
		return jerr.Get("error setting connection", err)
	}
	c := queue_pb.NewQueueClient(s.conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultSetTimeout)
	defer cancel()
	if _, err := c.DeleteMessages(ctx, &queue_pb.MessageUids{
		Topic: topic,
		Uids:  uids,
	}); err != nil {
		return jerr.Get("error deleting items for topics", err)
	}
	return nil
}

func NewClient(host string) *Client {
	return &Client{
		host: host,
	}
}
