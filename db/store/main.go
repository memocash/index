package store

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/config"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultOpenFilesCacheCapacity = 256
)

var conns = make(map[string]*leveldb.DB)

func getDb(topic string, shard uint) (*leveldb.DB, error) {
	if conns[topic] == nil {
		filename := GetDbFile(topic, shard)
		err := os.MkdirAll(filepath.Dir(filename), os.ModePerm)
		if err != nil {
			return nil, jerr.Get("error creating file directory", err)
		}
		openFilesCacheCapacity := config.GetOpenFilesCacheCapacity()
		if openFilesCacheCapacity == 0 {
			openFilesCacheCapacity = DefaultOpenFilesCacheCapacity
		}
		db, err := leveldb.OpenFile(filename, &opt.Options{
			OpenFilesCacheCapacity: openFilesCacheCapacity,
		})
		if err != nil {
			return nil, jerr.Get("error opening level db", err)
		}
		conns[topic] = db
	}
	return conns[topic], nil
}

func GetDbPrefix() string {
	prefix := config.GetDataPrefix()
	if prefix != "" {
		prefix = strings.TrimRight(prefix, string(os.PathSeparator)) + string(os.PathSeparator)
	}
	return prefix
}

func GetDbDir(shard uint) string {
	return "data/" + GetDbPrefix() + config.GetShardConfig(uint32(shard), config.GetQueueShards()).String()
}

func GetDbFile(topic string, shard uint) string {
	return GetDbDir(shard) + "/" + topic + ".ldb"
}

func GetMessageCount(topic string, shard uint) (int64, error) {
	db, err := getDb(topic, shard)
	if err != nil {
		return 0, jerr.Get("error getting db", err)
	}
	sizes, err := db.SizeOf([]util.Range{{
		Start: []byte{0x00},
		Limit: []byte{0xff},
	}})
	if err != nil {
		return 0, jerr.Get("error getting size of range", err)
	}
	if len(sizes) != 1 {
		return 0, jerr.Newf("error unexpected range slice len: %d", len(sizes))
	}
	return sizes[0], nil
}

func GetMessageCountReal(topic string, shard uint, prefix []byte) (uint64, error) {
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

const notFoundErrorMessage = "leveldb: not found"

func IsNotFoundError(err error) bool {
	return jerr.HasErrorPart(err, notFoundErrorMessage)
}
