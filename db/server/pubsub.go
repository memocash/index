package server

import (
	"bytes"
	"context"
	"github.com/jchavannes/jgo/jerr"
	"sync"
)

type Subscribe struct {
	Id       int64
	Shard    uint
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

var subscriberMutex sync.Mutex

func (s *PubSub) Subscribe(shard uint, topic string, start []byte, prefixes [][]byte) *Subscribe {
	//prefixStrings := jutil.ByteSliceStrings(prefixes)
	//jlog.Logf("New subscribe item shard: %d, topic: %s, start: %x, prefixes: %s\n",
	//	shard, topic, start, strings.Join(prefixStrings, " "))
	subscriberMutex.Lock()
	s.Incr++
	var sub = &Subscribe{
		Id:       s.Incr,
		Shard:    shard,
		Topic:    topic,
		Start:    start,
		Prefixes: prefixes,
		UidChan:  make(chan []byte),
		PubSub:   s,
	}
	s.Subs[sub.Id] = sub
	subscriberMutex.Unlock()
	return sub
}

func (s *PubSub) Close(id int64) {
	close(s.Subs[id].UidChan)
	delete(s.Subs, id)
}

func (s *PubSub) Publish(shard uint, topic string, uid []byte) {
	//jlog.Logf("New published item shard: %d, topic: %s, uid: %x\n", shard, topic, uid)
	for id := range s.Subs {
		var sub = s.Subs[id]
		if sub.Shard != shard || sub.Topic != topic {
			continue
		} else if len(sub.Start) > 0 {
			if bytes.Compare(uid, sub.Start) == 1 {
				sub.UidChan <- uid
			}
		} else if len(sub.Prefixes) > 0 {
			lenUid := len(uid)
			for _, prefix := range sub.Prefixes {
				lenPrefix := len(prefix)
				if lenPrefix <= lenUid && bytes.Equal(prefix, uid[:lenPrefix]) {
					sub.UidChan <- uid
					continue
				}
			}
		} else {
			sub.UidChan <- uid
		}
	}
}

var _globalPubSub *PubSub

func initNewListener() {
	if _globalPubSub == nil {
		_globalPubSub = &PubSub{
			Subs: make(map[int64]*Subscribe),
		}
	}
}

// ListenSingle returns nil if a matching new item is found, otherwise an error
func ListenSingle(ctx context.Context, shard uint, topic string, start []byte, prefixes [][]byte) error {
	initNewListener()
	var done = make(chan error)
	go func() {
		sub := _globalPubSub.Subscribe(shard, topic, start, prefixes)
		defer sub.Close()
		select {
		case <-ctx.Done():
			done <- jerr.Newf("error timeout listen single context")
		case <-sub.UidChan:
			done <- nil
		}
	}()
	return <-done
}

// Listen returns a channel of messages
func Listen(ctx context.Context, shard uint, topic string, prefixes [][]byte) chan []byte {
	initNewListener()
	var uidChan = make(chan []byte)
	go func() {
		sub := _globalPubSub.Subscribe(shard, topic, nil, prefixes)
		defer sub.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case uid := <-sub.UidChan:
				uidChan <- uid
			}
		}
	}()
	return uidChan
}

func ReceiveNew(shard uint, topic string, uid []byte) {
	initNewListener()
	_globalPubSub.Publish(shard, topic, uid)
}
