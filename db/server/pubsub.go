package server

import (
	"bytes"
	"context"
	"github.com/jchavannes/jgo/jerr"
)

type Subscribe struct {
	Topic    string
	Start    []byte
	Prefixes [][]byte
	Done     chan error
}

type PubSub struct {
	Subs []*Subscribe
}

func (s *PubSub) Subscribe(topic string, start []byte, prefixes [][]byte) *Subscribe {
	var sub = &Subscribe{
		Topic:    topic,
		Start:    start,
		Prefixes: prefixes,
		Done:     make(chan error),
	}
	s.Subs = append(s.Subs, sub)
	return sub
}

func (s *PubSub) Publish(topic string, uid []byte) {
	for i := 0; i < len(s.Subs); i++ {
		if s.Subs[i].Topic != topic {
		} else if len(s.Subs[i].Start) > 0 && bytes.Compare(uid, s.Subs[i].Start) == 1 {
			goto Found
		} else {
			lenUid := len(uid)
			for _, prefix := range s.Subs[i].Prefixes {
				lenPrefix := len(prefix)
				if lenPrefix <= lenUid && bytes.Equal(prefix, uid[:lenPrefix]) {
					goto Found
				}
			}
		}
		continue
	Found:
		go func(sub *Subscribe) {
			sub.Done <- nil
		}(s.Subs[i])
		s.Subs = append(s.Subs[:i], s.Subs[i+1:]...)
		i--
	}
}

var _globalPubSub *PubSub

func initNewListener() {
	if _globalPubSub == nil {
		_globalPubSub = &PubSub{}
	}
}

func ListenNew(ctx context.Context, topic string, start []byte, prefixes [][]byte) chan error {
	initNewListener()
	var done = make(chan error)
	go func() {
		sub := _globalPubSub.Subscribe(topic, start, prefixes)
		select {
		case <-ctx.Done():
			done <- jerr.Newf("error timeout")
		case err := <-sub.Done:
			if err != nil {
				done <- jerr.Getf(err, "error listening for event (%s %x)", topic, start)
			} else {
				done <- nil
			}
		}
	}()
	return done
}

func ReceiveNew(topic string, uid []byte) {
	initNewListener()
	_globalPubSub.Publish(topic, uid)
}
