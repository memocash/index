package store

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/db/client"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
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
	err = db.Write(batch, nil)
	if err != nil {
		return jerr.Get("error writing items to level db", err)
	}
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
	snap, err := db.GetSnapshot()
	if err != nil {
		return nil, jerr.Get("error getting db snapshot", err)
	}
	defer snap.Release()
	var messages []*Message
	for i := range uids {
		value, err := snap.Get(uids[i], nil)
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
	return messages, nil
}

func GetMessages(topic string, shard uint, prefixes [][]byte, start []byte, max int, newest bool) ([]*Message, error) {
	db, err := getDb(topic, shard)
	if err != nil {
		return nil, jerr.Get("error getting db", err)
	}
	var isGetLast bool
	if client.IsMaxStart(start) {
		isGetLast = true
		start = nil
	}
	if max == 0 {
		max = client.DefaultLimit
	}
	if len(prefixes) == 0 {
		prefixes = append(prefixes, []byte{})
	}
	snap, err := db.GetSnapshot()
	if err != nil {
		return nil, jerr.Get("error getting db snapshot", err)
	}
	defer snap.Release()
	var prefix []byte
	defer func() {
		if r := recover(); r != nil {
			jlog.Logf("prefix: %x, start: %x\n", prefix, start)
			panic(r)
		}
	}()
	var messages []*Message
	for _, prefix = range prefixes {
		var iter iterator.Iterator
		if newest {
			var iterRange *util.Range
			if len(start) > 0 {
				iterRange = &util.Range{
					Limit: start,
				}
			}
			iter = snap.NewIterator(iterRange, nil)
		} else {
			iterRange := util.BytesPrefix(prefix)
			if len(start) > 0 && bytes.Compare(start, prefix) != -1 {
				iterRange.Start = start
			}
			iter = snap.NewIterator(iterRange, nil)
		}
		if isGetLast {
			if iter.Last() {
				messages = append(messages, &Message{
					Uid:     GetPtrSlice(iter.Key()),
					Message: GetPtrSlice(iter.Value()),
				})
			}
		} else if newest {
			for ok := iter.Last(); ok; ok = iter.Prev() {
				messages = append(messages, &Message{
					Uid:     GetPtrSlice(iter.Key()),
					Message: GetPtrSlice(iter.Value()),
				})
				if len(messages) >= max {
					break
				}
			}
		} else {
			for iter.Next() {
				messages = append(messages, &Message{
					Uid:     GetPtrSlice(iter.Key()),
					Message: GetPtrSlice(iter.Value()),
				})
				if len(messages) >= max {
					break
				}
			}
		}
		iter.Release()
		err = iter.Error()
		if err != nil {
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
