package store

import (
	"bytes"
	"fmt"
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

type RequestByPrefixes struct {
	Topic    string // required
	Shard    uint   // required
	Prefixes []Prefix
	Max      int
	Newest   bool
}

type Prefix struct {
	Prefix []byte
	Start  []byte
	Max    int
}

// GetByPrefixes returns messages.
func GetByPrefixes(request RequestByPrefixes) ([]*Message, error) {
	db, err := getDb(request.Topic, request.Shard)
	if err != nil {
		return nil, fmt.Errorf("error getting db shard %d; %w", request.Shard, err)
	}

	var maxResults = request.Max
	if maxResults == 0 {
		maxResults = client.HugeLimit
	}

	var prefixes = request.Prefixes
	if len(prefixes) == 0 {
		prefixes = append(prefixes, Prefix{})
	}

	var messages []*Message
	defer func() {
		metric.AddTopicRead(metric.TopicRead{
			Topic:    request.Topic,
			Quantity: len(messages),
		})
	}()

	for _, prefix := range prefixes {
		prefixMessages, err := getPrefixMessages(db, prefix, request.Newest, maxResults-len(messages))
		if err != nil {
			return nil, fmt.Errorf("error getting prefix messages; %w", err)
		}

		messages = append(messages, prefixMessages...)

		if len(messages) >= maxResults {
			break
		}
	}

	return messages, nil
}

func getPrefixMessages(db *leveldb.DB, prefix Prefix, newest bool, totalMaxLeft int) ([]*Message, error) {
	var maxPrefixResults = prefix.Max
	if maxPrefixResults == 0 {
		maxPrefixResults = client.HugeLimit
	}
	if maxPrefixResults > totalMaxLeft {
		maxPrefixResults = totalMaxLeft
	}

	iterRange := util.BytesPrefix(prefix.Prefix)
	if len(prefix.Start) > 0 {
		if newest && (len(prefix.Prefix) == 0 || bytes.Compare(prefix.Start, prefix.Prefix) != 1) {
			iterRange.Limit = prefix.Start
		} else if !newest && (len(prefix.Prefix) == 0 || bytes.Compare(prefix.Start, prefix.Prefix) != -1) {
			iterRange.Start = prefix.Start
		}
	}

	iter := db.NewIterator(iterRange, nil)
	var prefixMessages []*Message

	storePrefixMessage := func() bool {
		prefixMessages = append(prefixMessages, &Message{
			Uid:     GetPtrSlice(iter.Key()),
			Message: GetPtrSlice(iter.Value()),
		})
		return len(prefixMessages) < maxPrefixResults
	}

	if newest {
		for ok := iter.Last(); ok && storePrefixMessage(); ok = iter.Prev() {
		}
	} else {
		for iter.Next() && storePrefixMessage() {
		}
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("error with releasing iterator; %w", err)
	}

	return prefixMessages, nil
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
