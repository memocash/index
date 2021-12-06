package server

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
)

func GetHost(port uint) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}

type Db interface {
	GetMsg(*RequestSingle) (*Msg, error)
	GetMsgs(*Request) ([]*Msg, error)
	SaveMsgs([]*Msg) error
	GetCount(*RequestCount) (uint64, error)
	GetTopicList() ([]TopicInfo, error)
	DeleteMessages(*RequestDelete) error
}

type Msg struct {
	Topic   string
	Uid     []byte
	Message []byte
}

func (m Msg) GetShard() uint {
	return client.GetByteShard(m.Uid)
}

type Request struct {
	Topic    string
	Prefixes [][]byte // Filter only these prefixes, e.g. prefix = "b". Ignore Start if more than 1, except overload
	Start    []byte   // Seek to this location to start, e.g. seek to "baltimore", can be overloaded to say last item
	Uids     [][]byte // When set, Start and Prefixes are ignored
	Max      uint32
	Wait     bool
	Newest   bool
	Context  context.Context
}

type RequestSingle struct {
	Topic string
	Uid   []byte
}

type RequestOld struct {
	Topic  string
	Offset uint64
	Max    uint32
	Wait   bool
}

type RequestDelete struct {
	Topic string
	Uids  [][]byte
}

type MsgDone struct {
	Msgs []*Msg
	Done chan error
}

func NewMsgDone(msgs []*Msg) *MsgDone {
	return &MsgDone{
		Msgs: msgs,
		Done: make(chan error),
	}
}

type TopicInfo struct {
	Name  string
	Count uint64
}

type RequestCount struct {
	Topic  string
	Prefix []byte
}
