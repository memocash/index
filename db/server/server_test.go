package server_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/memocash/index/db/proto/queue_pb"
	"github.com/memocash/index/db/server"
	"github.com/memocash/index/db/store"
	"github.com/syndtr/goleveldb/leveldb"
)

const testTopic = "server-test"

// TestSearchPartlessPattern pins that a protobuf-valid pattern with anchors
// but no parts — documented to match everything — flows through the handler
// as a full-range scan instead of panicking the serving process (an invalid
// client can send this shape remotely).
func TestSearchPartlessPattern(t *testing.T) {
	database, err := leveldb.OpenFile(filepath.Join(t.TempDir(), "db"), nil)
	if err != nil {
		t.Fatalf("error opening level db; %v", err)
	}
	store.SetConn(store.GetConnId(testTopic, 0), database)
	defer store.CloseAll()
	if err := store.SaveMessages(testTopic, 0, []*store.Message{
		{Uid: []byte("uid-1"), Message: []byte("message-1")},
		{Uid: []byte("uid-2"), Message: []byte("message-2")},
	}); err != nil {
		t.Fatalf("error saving messages; %v", err)
	}
	s := &server.Server{Shard: 0}
	for _, pattern := range []*queue_pb.Pattern{
		{},
		{AnchorStart: true},
		{AnchorEnd: true},
		{AnchorStart: true, AnchorEnd: true},
	} {
		reply, err := s.Search(context.Background(), &queue_pb.SearchRequest{
			Topic:    testTopic,
			Patterns: []*queue_pb.SearchPattern{{Uid: pattern}},
		})
		if err != nil {
			t.Fatalf("pattern %+v: error searching; %v", pattern, err)
		}
		if len(reply.Messages) != 2 {
			t.Errorf("pattern %+v: got %d messages, want 2", pattern, len(reply.Messages))
		}
	}
}
