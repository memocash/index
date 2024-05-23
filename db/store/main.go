package store

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/config"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	DefaultOpenFilesCacheCapacity = 256
)

var conns = make(map[string]*leveldb.DB)
var connsMutex = sync.RWMutex{}

func CloseAll() {
	connsMutex.Lock()
	defer connsMutex.Unlock()
	for k, db := range conns {
		db.Close()
		delete(conns, k)
	}
}

func getDb(topic string, shard uint) (*leveldb.DB, error) {
	connsMutex.Lock()
	defer connsMutex.Unlock()
	connId := fmt.Sprintf("%d:%s", shard, topic)
	if conns[connId] == nil {
		if conns[connId] != nil {
			return conns[connId], nil
		}
		filename := GetDbFile(topic, shard)
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			return nil, fmt.Errorf("error creating file directory; %w", err)
		}
		openFilesCacheCapacity := config.GetOpenFilesCacheCapacity()
		if openFilesCacheCapacity == 0 {
			openFilesCacheCapacity = DefaultOpenFilesCacheCapacity
		}
		compactionDataSize := config.GetCompactionDataSize()
		db, err := leveldb.OpenFile(filename, &opt.Options{
			OpenFilesCacheCapacity: openFilesCacheCapacity,
			CompactionTableSize:    compactionDataSize * opt.MiB,
			WriteBuffer:            compactionDataSize * 2 * opt.MiB,
		})
		if err != nil {
			return nil, fmt.Errorf("error opening level db: %s; %w", filename, err)
		}
		conns[connId] = db
	}
	return conns[connId], nil
}

func GetDbPrefix() string {
	prefix := config.GetDataPrefix()
	if prefix != "" {
		prefix = strings.TrimRight(prefix, string(os.PathSeparator)) + string(os.PathSeparator)
	}
	return prefix
}

func GetDataDir() string {
	dataDir := config.GetDataDir()
	if dataDir == "" {
		dataDir = "data"
	}
	return dataDir
}

func GetDbDir(shard uint) string {
	return GetDataDir() + "/" + GetDbPrefix() + config.GetShardConfig(uint32(shard), config.GetQueueShards()).String()
}

func GetDbFile(topic string, shard uint) string {
	return GetDbDir(shard) + "/" + topic + ".ldb"
}

func GetMessageCount(topic string, shard uint) (int64, error) {
	db, err := getDb(topic, shard)
	if err != nil {
		return 0, fmt.Errorf("error getting db; %w", err)
	}
	sizes, err := db.SizeOf([]util.Range{{
		Start: []byte{0x00},
		Limit: []byte{0xff},
	}})
	if err != nil {
		return 0, fmt.Errorf("error getting size of range; %w", err)
	}
	if len(sizes) != 1 {
		return 0, fmt.Errorf("error unexpected range slice len: %d", len(sizes))
	}
	return sizes[0], nil
}

func GetMessageCountReal(topic string, shard uint, prefix []byte) (uint64, error) {
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

const notFoundErrorMessage = "leveldb: not found"

func IsNotFoundError(err error) bool {
	return jerr.HasErrorPart(err, notFoundErrorMessage)
}
