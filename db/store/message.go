package store

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
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
		return jerr.Get("error getting level db", err)
	}
	batch := new(leveldb.Batch)
	for _, message := range messages {
		batch.Put(message.Uid, message.Message)
	}
	if err = db.Write(batch, nil); err != nil {
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
		return nil, jerr.Getf(err, "error getting db shard %d", shard)
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
			jlog.Logf("PANIC prefix: %x, start: %x\n", prefix, start)
			panic(r)
		}
	}()
	var messages []*Message
	sort.Slice(prefixes, func(i, j int) bool {
		return jutil.ByteLT(prefixes[i], prefixes[j])
	})
	if !newest && !isGetLast {
		iter := snap.NewIterator(nil, nil)
		defer iter.Release()
		for _, prefix = range prefixes {
			var prefixMessages []*Message
			var seek = start
			if len(seek) == 0 || jutil.ByteLT(start, prefix) {
				seek = prefix
			}
			if !iter.Seek(seek) {
				continue
			}
			for ok := true; ok; ok = iter.Next() {
				uid := GetPtrSlice(iter.Key())
				if !jutil.HasPrefix(uid, prefix) {
					break
				}
				prefixMessages = append(prefixMessages, &Message{
					Uid:     uid,
					Message: GetPtrSlice(iter.Value()),
				})
				if len(prefixMessages) >= max {
					break
				}
			}
			messages = append(messages, prefixMessages...)
		}
		if err = iter.Error(); err != nil {
			return nil, jerr.Get("error with iterator", err)
		}
		return messages, nil
	}
	for _, prefix = range prefixes {
		var prefixMessages []*Message
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
				prefixMessages = append(prefixMessages, &Message{
					Uid:     GetPtrSlice(iter.Key()),
					Message: GetPtrSlice(iter.Value()),
				})
			}
		} else if newest {
			for ok := iter.Last(); ok; ok = iter.Prev() {
				prefixMessages = append(prefixMessages, &Message{
					Uid:     GetPtrSlice(iter.Key()),
					Message: GetPtrSlice(iter.Value()),
				})
				if len(prefixMessages) >= max {
					break
				}
			}
		} else {
			for iter.Next() {
				prefixMessages = append(prefixMessages, &Message{
					Uid:     GetPtrSlice(iter.Key()),
					Message: GetPtrSlice(iter.Value()),
				})
				if len(prefixMessages) >= max {
					break
				}
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
