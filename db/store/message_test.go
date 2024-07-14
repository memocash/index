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
	testDbPath := filepath.Join(os.TempDir(), fmt.Sprintf("goleveldbtest-%d", os.Getuid()))
	if err := os.RemoveAll(testDbPath); err != nil {
		return fmt.Errorf("error removing old db; %w", err)
	}

	db, err := leveldb.OpenFile(testDbPath, nil)
	if err != nil {
		return fmt.Errorf("error opening level db; %w", err)
	}

	store.SetConn(store.GetConnId(TestTopic, TestShard), db)

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

func TestGetByPrefixes(t *testing.T) {
	if err := initTestDb(); err != nil {
		t.Errorf("error initializing test db; %v", err)
	}

	const MessageCount = 10
	for i := 0; i < MessageCount; i++ {
		if err := store.SaveMessages(TestTopic, TestShard, []*store.Message{{
			Uid:     []byte(fmt.Sprintf("test-%d", i)),
			Message: []byte(fmt.Sprintf("test-message-%d", i)),
		}}); err != nil {
			t.Errorf("error saving prefix messages; %v", err)
		}
	}

	prefix := []byte("test-")
	messages, err := store.GetMessages(TestTopic, TestShard, [][]byte{prefix}, nil, 0, false)
	if err != nil {
		t.Errorf("error getting message; %v", err)
		return
	}

	if len(messages) != MessageCount {
		t.Errorf("unexpected number of messages: %d, expected %d\n", len(messages), MessageCount)
		return
	}

	for i := range messages {
		message := messages[i]
		expectedUid := fmt.Sprintf("test-%d", i)
		if string(message.Uid) != expectedUid {
			t.Errorf("unexpected message uid: %s, expected %s\n", message.Uid, expectedUid)
			return
		}
		expectedMessage := fmt.Sprintf("test-message-%d", i)
		if string(message.Message) != expectedMessage {
			t.Errorf("unexpected message: %s, expected %s\n", message.Message, expectedMessage)
			return
		}
	}
}
