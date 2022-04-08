package server

import (
	"bytes"
	"context"
	"github.com/jchavannes/jgo/jerr"
)

type Subscribe struct {
	Id       int64
	Topic    string
	Start    []byte
	Prefixes [][]byte
	UidChan  chan []byte
	PubSub   *PubSub
}

func (s *Subscribe) Close() {
	s.PubSub.Close(s.Id)
}

type PubSub struct {
	Incr int64
	Subs map[int64]*Subscribe
}

func (s *PubSub) Subscribe(topic string, start []byte, prefixes [][]byte) *Subscribe {
	s.Incr++
	var sub = &Subscribe{
		Id:       s.Incr,
		Topic:    topic,
		Start:    start,
		Prefixes: prefixes,
		UidChan:  make(chan []byte),
		PubSub:   s,
	}
	s.Subs[sub.Id] = sub
	return sub
}

func (s *PubSub) Close(id int64) {
	close(s.Subs[id].UidChan)
	delete(s.Subs, id)
}

func (s *PubSub) Publish(topic string, uid []byte) {
	for id := range s.Subs {
		var sub = s.Subs[id]
		if sub.Topic != topic {
		} else if len(sub.Start) > 0 && bytes.Compare(uid, sub.Start) == 1 {
			sub.UidChan <- uid
		} else {
			lenUid := len(uid)
			for _, prefix := range sub.Prefixes {
				lenPrefix := len(prefix)
				if lenPrefix <= lenUid && bytes.Equal(prefix, uid[:lenPrefix]) {
					sub.UidChan <- uid
					continue
				}
			}
		}
	}
}

var _globalPubSub *PubSub

func initNewListener() {
	if _globalPubSub == nil {
		_globalPubSub = &PubSub{}
	}
}

// ListenSingle returns nil if a matching new item is found, otherwise an error
func ListenSingle(ctx context.Context, topic string, start []byte, prefixes [][]byte) error {
	initNewListener()
	var done = make(chan error)
	go func() {
		sub := _globalPubSub.Subscribe(topic, start, prefixes)
		defer sub.Close()
		select {
		case <-ctx.Done():
			done <- jerr.Newf("error timeout")
		case <-sub.UidChan:
			done <- nil
		}
	}()
	return <-done
}

// Listen returns a channel of messages
func Listen(ctx context.Context, topic string, prefixes [][]byte) chan []byte {
	initNewListener()
	var uidChan = make(chan []byte)
	go func() {
		sub := _globalPubSub.Subscribe(topic, nil, prefixes)
		defer sub.Close()
		for {
			select {
			case <-ctx.Done():
				break
			case uid := <-sub.UidChan:
				uidChan <- uid
			}
		}
	}()
	return uidChan
}

func ReceiveNew(topic string, uid []byte) {
	initNewListener()
	_globalPubSub.Publish(topic, uid)
}
