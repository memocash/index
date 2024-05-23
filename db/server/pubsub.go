package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/memocash/index/db/metric"
	"sync"
	"time"
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
	Incr  int64
	Subs  map[int64]*Subscribe
	Mutex sync.Mutex
}

func (s *PubSub) Subscribe(shard uint, topic string, start []byte, prefixes [][]byte) *Subscribe {
	//prefixStrings := jutil.ByteSliceStrings(prefixes)
	//log.Printf("New subscribe item shard: %d, topic: %s, start: %x, prefixes: %s\n",
	//	shard, topic, start, strings.Join(prefixStrings, " "))
	var sub = &Subscribe{
		Shard:    shard,
		Topic:    topic,
		Start:    start,
		Prefixes: prefixes,
		UidChan:  make(chan []byte),
		PubSub:   s,
	}
	s.Mutex.Lock()
	s.Incr++
	sub.Id = s.Incr
	s.Subs[sub.Id] = sub
	s.Mutex.Unlock()
	return sub
}

func (s *PubSub) Close(id int64) {
	s.Mutex.Lock()
	close(s.Subs[id].UidChan)
	delete(s.Subs, id)
	s.Mutex.Unlock()
}

func (s *PubSub) Publish(shard uint, topic string, uid []byte) {
	//log.Printf("New published item shard: %d, topic: %s, uid: %x\n", shard, topic, uid)
	s.Mutex.Lock()
	for id := range s.Subs {
		var sub = s.Subs[id]
		s.Mutex.Unlock()
		if sub.Shard != shard || sub.Topic != topic {
			goto End
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
					goto End
				}
			}
		} else {
			sub.UidChan <- uid
		}
	End:
		s.Mutex.Lock()
	}
	s.Mutex.Unlock()
}

var _globalPubSub *PubSub

func initNewListener() {
	if _globalPubSub == nil {
		_globalPubSub = &PubSub{
			Subs: make(map[int64]*Subscribe),
		}
		go func() {
			t := time.NewTicker(10 * time.Second)
			for {
				<-t.C
				_globalPubSub.Mutex.Lock()
				quantity := len(_globalPubSub.Subs)
				_globalPubSub.Mutex.Unlock()
				metric.AddListenCount(metric.ListenCount{Quantity: quantity})
			}
		}()
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
			done <- fmt.Errorf("error timeout listen single context")
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
		defer func() {
			sub.Close()
			close(uidChan)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case uid, ok := <-sub.UidChan:
				if !ok {
					return
				}
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
