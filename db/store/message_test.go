package store_test

import (
	"bytes"
	"fmt"
	"github.com/memocash/index/db/store"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"testing"
)

const TestTopic = "test"
const TestShard = 0

const PrefixTest = "test"
const PrefixOther = "other"

var (
	testMessageTest0 = &store.Message{Uid: []byte("test-0"), Message: []byte("message-0")}
	testMessageTest1 = &store.Message{Uid: []byte("test-1"), Message: []byte("message-1")}
	testMessageTest2 = &store.Message{Uid: []byte("test-2"), Message: []byte("message-2")}
	testMessageTest3 = &store.Message{Uid: []byte("test-3"), Message: []byte("message-3")}
	testMessageTest4 = &store.Message{Uid: []byte("test-4"), Message: []byte("message-4")}
	testMessageTest5 = &store.Message{Uid: []byte("test-5"), Message: []byte("message-5")}
	testMessageTest6 = &store.Message{Uid: []byte("test-6"), Message: []byte("message-6")}
	testMessageTest7 = &store.Message{Uid: []byte("test-7"), Message: []byte("message-7")}
	testMessageTest8 = &store.Message{Uid: []byte("test-8"), Message: []byte("message-8")}
	testMessageTest9 = &store.Message{Uid: []byte("test-9"), Message: []byte("message-9")}

	testMessageOther0 = &store.Message{Uid: []byte("other-0"), Message: []byte("message-0")}
	testMessageOther1 = &store.Message{Uid: []byte("other-1"), Message: []byte("message-1")}
	testMessageOther2 = &store.Message{Uid: []byte("other-2"), Message: []byte("message-2")}
	testMessageOther3 = &store.Message{Uid: []byte("other-3"), Message: []byte("message-3")}
	testMessageOther4 = &store.Message{Uid: []byte("other-4"), Message: []byte("message-4")}
	testMessageOther5 = &store.Message{Uid: []byte("other-5"), Message: []byte("message-5")}
	testMessageOther6 = &store.Message{Uid: []byte("other-6"), Message: []byte("message-6")}
	testMessageOther7 = &store.Message{Uid: []byte("other-7"), Message: []byte("message-7")}
	testMessageOther8 = &store.Message{Uid: []byte("other-8"), Message: []byte("message-8")}
	testMessageOther9 = &store.Message{Uid: []byte("other-9"), Message: []byte("message-9")}
)

type GetMessagesTest struct {
	Name     string
	Prefixes []store.Prefix
	Limit    int
	Desc     bool
	Expected []*store.Message
}

var tests = []GetMessagesTest{{
	Name:     "Basic",
	Prefixes: []store.Prefix{{Prefix: []byte(PrefixTest), Max: 5}, {Prefix: []byte(PrefixOther), Max: 5}},
	Expected: []*store.Message{
		testMessageTest0, testMessageTest1, testMessageTest2, testMessageTest3, testMessageTest4,
		testMessageOther0, testMessageOther1, testMessageOther2, testMessageOther3, testMessageOther4,
	},
}, {
	Name: "Start 1",
	Prefixes: []store.Prefix{
		{Prefix: []byte(PrefixTest), Start: []byte(fmt.Sprintf("%s-%d", PrefixOther, 1)), Max: 5},
		{Prefix: []byte(PrefixOther), Start: []byte(fmt.Sprintf("%s-%d", PrefixOther, 1)), Max: 5},
	},
	Expected: []*store.Message{
		testMessageTest0, testMessageTest1, testMessageTest2, testMessageTest3, testMessageTest4,
		testMessageOther1, testMessageOther2, testMessageOther3, testMessageOther4, testMessageOther5,
	},
}, {
	Name: "Start 2",
	Prefixes: []store.Prefix{
		{Prefix: []byte(PrefixTest), Start: []byte(fmt.Sprintf("%s-%d", PrefixTest, 1))},
		{Prefix: []byte(PrefixOther), Start: []byte(fmt.Sprintf("%s-%d", PrefixTest, 1))},
	},
	Limit: 5,
	Expected: []*store.Message{
		testMessageTest1, testMessageTest2, testMessageTest3, testMessageTest4, testMessageTest5,
	},
}, {
	Name:     "Descending",
	Prefixes: []store.Prefix{{Prefix: []byte(PrefixTest), Max: 5}, {Prefix: []byte(PrefixOther), Max: 5}},
	Desc:     true,
	Expected: []*store.Message{
		testMessageTest9, testMessageTest8, testMessageTest7, testMessageTest6, testMessageTest5,
		testMessageOther9, testMessageOther8, testMessageOther7, testMessageOther6, testMessageOther5,
	},
}}

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

	if err := store.SaveMessages(TestTopic, TestShard, []*store.Message{
		testMessageTest0, testMessageTest1, testMessageTest2, testMessageTest3, testMessageTest4,
		testMessageTest5, testMessageTest6, testMessageTest7, testMessageTest8, testMessageTest9,
		testMessageOther0, testMessageOther1, testMessageOther2, testMessageOther3, testMessageOther4,
		testMessageOther5, testMessageOther6, testMessageOther7, testMessageOther8, testMessageOther9,
	}); err != nil {
		return fmt.Errorf("error saving prefix messages; %w", err)
	}

	return nil
}

func TestGetMessage(t *testing.T) {
	if err := initTestDb(); err != nil {
		t.Errorf("error initializing test db; %v", err)
	}
	defer store.CloseAll()

	message, err := store.GetMessage(TestTopic, TestShard, testMessageTest1.Uid)
	if err != nil {
		t.Errorf("error getting message; %v", err)
		return
	}

	if message == nil {
		t.Errorf("message not found")
		return
	}

	if !bytes.Equal(message.Message, testMessageTest1.Message) {
		t.Errorf("message not correct")
		return
	}
}

func TestGetByPrefixes(t *testing.T) {
	if err := initTestDb(); err != nil {
		t.Errorf("error initializing test db; %v", err)
	}
	defer store.CloseAll()

	for _, test := range tests {
		messages, err := store.GetByPrefixes(store.RequestByPrefixes{
			Topic:    TestTopic,
			Shard:    TestShard,
			Prefixes: test.Prefixes,
			Limit:    test.Limit,
			Desc:     test.Desc,
		})
		if err != nil {
			t.Errorf("%s test error getting message; %v", test.Name, err)
			return
		}

		if len(messages) != len(test.Expected) {
			t.Errorf("%s test error unexpected number of messages: %d, expected %d\n",
				test.Name, len(messages), len(test.Expected))
			return
		}

		for i := range messages {
			message := messages[i]
			expected := test.Expected[i]

			if !bytes.Equal(message.Uid, expected.Uid) {
				t.Errorf("%s test error unexpected message uid: %s, expected %s\n",
					test.Name, message.Uid, expected.Uid)
				return
			}

			if !bytes.Equal(message.Message, expected.Message) {
				t.Errorf("%s test error unexpected message: %s, expected %s\n",
					test.Name, message.Message, expected.Message)
				return
			}
		}
	}
}
