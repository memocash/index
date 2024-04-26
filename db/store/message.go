package store

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/metric"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Message struct {
	Uid     []byte
	Message []byte
}

func SaveMessages(topic string, shard uint, messages []*Message) error {
	db, err := getDb(topic, shard)
	if err != nil {
		return jerr.Get("error getting level db", err)
	}
	batch := new(leveldb.Batch)
	for _, message := range messages {
		batch.Put(message.Uid, message.Message)
	}
	if err = db.Write(batch, nil); err != nil {
		return jerr.Get("error writing items to level db", err)
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
		return nil, jerr.Get("error getting db", err)
	}
	value, err := db.Get(uid, nil)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, nil
		}
		return nil, jerr.Get("error getting message", err)
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
		return nil, jerr.Get("error getting db", err)
	}
	var messages []*Message
	for i := range uids {
		value, err := db.Get(uids[i], nil)
		if err != nil {
			if IsNotFoundError(err) {
				continue
			}
			return nil, jerr.Getf(err, "error getting message: %x", uids[i])
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
		return nil, jerr.Getf(err, "error getting db shard %d", shard)
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
			return nil, jerr.Get("error with releasing iterator", err)
		}
	}
	return messages, nil
}

func DeleteMessages(topic string, shard uint, uids [][]byte) error {
	db, err := getDb(topic, shard)
	if err != nil {
		return jerr.Get("error getting level db for topic", err)
	}
	batch := new(leveldb.Batch)
	for _, uid := range uids {
		batch.Delete(uid)
	}
	err = db.Write(batch, nil)
	if err != nil {
		return jerr.Get("error batch deleting items in level db", err)
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
		return 0, jerr.Get("error getting db", err)
	}
	snap, err := db.GetSnapshot()
	if err != nil {
		return 0, jerr.Get("error getting db snapshot", err)
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
		return 0, jerr.Get("error with releasing iterator", err)
	}
	return count, nil
}
