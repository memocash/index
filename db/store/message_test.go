package store_test

import (
	"fmt"
	"github.com/memocash/index/db/store"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"testing"
)

const TestTopic = "test"
const TestShard = 0

func initTestDb() error {
	connId := store.GetConnId(TestTopic, TestShard)
	testDbPath := filepath.Join(os.TempDir(), fmt.Sprintf("goleveldbtest-%d", os.Getuid()))
	if err := os.RemoveAll(testDbPath); err != nil {
		return fmt.Errorf("error removing old db; %w", err)
	}
	db, err := leveldb.OpenFile(testDbPath, nil)
	if err != nil {
		return fmt.Errorf("error opening level db; %w", err)
	}
	store.SetConn(connId, db)
	return nil
}

func TestGetMessage(t *testing.T) {
	if err := initTestDb(); err != nil {
		t.Errorf("error initializing test db; %v", err)
	}
	if err := store.SaveMessages(TestTopic, TestShard, []*store.Message{{
		Uid:     []byte("test-uid"),
		Message: []byte("test-message"),
	}}); err != nil {
		t.Errorf("error saving message; %v", err)
	}
	message, err := store.GetMessage(TestTopic, TestShard, []byte("test-uid"))
	if err != nil {
		t.Errorf("error getting message; %v", err)
		return
	}
	if message == nil {
		t.Errorf("message not found")
		return
	}
	if string(message.Message) != "test-message" {
		t.Errorf("message not correct")
		return
	}
}
