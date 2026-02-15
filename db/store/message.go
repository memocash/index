package store

import (
	"bytes"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/metric"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sort"
)

type Message struct {
	Uid     []byte
	Message []byte
}

func SaveMessages(topic string, shard uint, messages []*Message) error {
	db, err := getDb(topic, shard)
	if err != nil {
		return fmt.Errorf("error getting level db; %w", err)
	}
	batch := new(leveldb.Batch)
	for _, message := range messages {
		batch.Put(message.Uid, message.Message)
	}
	if err = db.Write(batch, nil); err != nil {
		return fmt.Errorf("error writing items to level db; %w", err)
	}
	metric.AddTopicSave(metric.TopicSave{
		Topic:    topic,
		Quantity: len(messages),
	})
	return nil
}

func GetMessage(topic string, shard uint, uid []byte) (*Message, error) {
	db, err := getDb(topic, shard)
	if err != nil {
		return nil, fmt.Errorf("error getting db; %w", err)
	}
	value, err := db.Get(uid, nil)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting message; %w", err)
	}
	metric.AddTopicRead(metric.TopicRead{
		Topic:    topic,
		Quantity: 1,
	})
	return &Message{
		Uid:     uid,
		Message: value,
	}, nil
}

func GetMessagesByUids(topic string, shard uint, uids [][]byte) ([]*Message, error) {
	db, err := getDb(topic, shard)
	if err != nil {
		return nil, fmt.Errorf("error getting db; %w", err)
	}
	var messages []*Message
	for i := range uids {
		value, err := db.Get(uids[i], nil)
		if err != nil {
			if IsNotFoundError(err) {
				continue
			}
			return nil, fmt.Errorf("error getting message: %x; %w", uids[i], err)
		}
		messages = append(messages, &Message{
			Uid:     uids[i],
			Message: value,
		})
	}
	metric.AddTopicRead(metric.TopicRead{
		Topic:    topic,
		Quantity: len(messages),
	})
	return messages, nil
}

// GetMessages returns messages. Options:
//   - Topic: required
//   - Shard: required
//   - Prefixes: optional, limits results
//   - Start: optional, where to start
//   - Max: optional, number of results
//   - Newest: optional, reverse order (fff first, instead of 000)
func GetMessages(topic string, shard uint, prefixes [][]byte, start []byte, max int, newest bool) ([]*Message, error) {
	db, err := getDb(topic, shard)
	if err != nil {
		return nil, fmt.Errorf("error getting db shard %d; %w", shard, err)
	}
	if max == 0 {
		max = client.HugeLimit
	}
	if len(prefixes) == 0 {
		prefixes = append(prefixes, []byte{})
	}
	var messages []*Message
	defer func() {
		metric.AddTopicRead(metric.TopicRead{
			Topic:    topic,
			Quantity: len(messages),
		})
	}()
	if !newest && len(start) == 0 && len(prefixes) > 1 {
		return getMessagesSorted(db, prefixes, max)
	}
	for _, prefix := range prefixes {
		var prefixMessages []*Message
		iterRange := util.BytesPrefix(prefix)
		if len(start) > 0 {
			if newest && (len(prefix) == 0 || bytes.Compare(start, prefix) != 1) {
				iterRange.Limit = start
			} else if !newest && (len(prefix) == 0 || bytes.Compare(start, prefix) != -1) {
				iterRange.Start = start
			}
		}
		iter := db.NewIterator(iterRange, nil)
		storePrefixMessage := func() bool {
			prefixMessages = append(prefixMessages, &Message{
				Uid:     GetPtrSlice(iter.Key()),
				Message: GetPtrSlice(iter.Value()),
			})
			return len(prefixMessages) < max
		}
		if newest {
			for ok := iter.Last(); ok && storePrefixMessage(); ok = iter.Prev() {
			}
		} else {
			for iter.Next() && storePrefixMessage() {
			}
		}
		messages = append(messages, prefixMessages...)
		iter.Release()
		if err = iter.Error(); err != nil {
			return nil, fmt.Errorf("error with releasing iterator; %w", err)
		}
	}
	return messages, nil
}

func getMessagesSorted(db *leveldb.DB, prefixes [][]byte, max int) ([]*Message, error) {
	sorted := make([][]byte, len(prefixes))
	copy(sorted, prefixes)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i], sorted[j]) < 0
	})
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	var messages []*Message
	for _, prefix := range sorted {
		found := iter.Seek(prefix)
		for count := 0; found && bytes.HasPrefix(iter.Key(), prefix) && count < max; found = iter.Next() {
			messages = append(messages, &Message{
				Uid:     GetPtrSlice(iter.Key()),
				Message: GetPtrSlice(iter.Value()),
			})
			count++
		}
	}
	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("error with sorted iterator; %w", err)
	}
	return messages, nil
}

func DeleteMessages(topic string, shard uint, uids [][]byte) error {
	db, err := getDb(topic, shard)
	if err != nil {
		return fmt.Errorf("error getting level db for topic; %w", err)
	}
	batch := new(leveldb.Batch)
	for _, uid := range uids {
		batch.Delete(uid)
	}
	err = db.Write(batch, nil)
	if err != nil {
		return fmt.Errorf("error batch deleting items in level db; %w", err)
	}
	return nil
}

func GetPtrSlice(s []byte) []byte {
	var d = make([]byte, len(s))
	copy(d, s)
	return d
}

func GetCount(topic string, prefix []byte, shard uint) (uint64, error) {
	db, err := getDb(topic, shard)
	if err != nil {
		return 0, fmt.Errorf("error getting db; %w", err)
	}
	snap, err := db.GetSnapshot()
	if err != nil {
		return 0, fmt.Errorf("error getting db snapshot; %w", err)
	}
	defer snap.Release()
	iterRange := util.BytesPrefix(prefix)
	iter := snap.NewIterator(iterRange, nil)
	var count uint64
	for iter.Next() {
		count++
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return 0, fmt.Errorf("error with releasing iterator; %w", err)
	}
	return count, nil
}
