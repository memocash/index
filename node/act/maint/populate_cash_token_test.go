package maint_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/proto/queue_pb"
	"github.com/memocash/index/node/act/maint"
	"github.com/memocash/index/ref/config"
	"google.golang.org/grpc"
)

// TestPopulateCashTokenAllShardsFail points every configured shard at a
// queue server whose RPCs all fail and requires Populate to report that
// failure with a live parent context. An earlier version arbitrated worker
// errors and worker completion in one select, which could pick completion
// first and return success for a backfill that did nothing; recording the
// error in the worker makes the outcome deterministic, and the repeated
// attempts guard against that regression.
func TestPopulateCashTokenAllShardsFail(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("error listening; %v", err)
	}
	server := grpc.NewServer()
	queue_pb.RegisterQueueServer(server, &queue_pb.UnimplementedQueueServer{})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	shards := config.GetQueueShards()
	if len(shards) == 0 {
		t.Fatal("error no queue shards configured")
	}
	original := append([]config.Shard(nil), shards...)
	defer func() {
		copy(shards, original)
		client.ResetConnections()
	}()
	for i := range shards {
		shards[i].Host = "127.0.0.1"
		shards[i].Port = listener.Addr().(*net.TCPAddr).Port
	}
	client.ResetConnections()
	for attempt := 0; attempt < 200; attempt++ {
		populate := maint.NewPopulateCashToken(context.Background())
		populate.LogInterval = time.Millisecond
		err := populate.Populate(true)
		if err == nil {
			t.Fatalf("attempt %d: Populate returned nil although every shard RPC fails", attempt)
		}
		if populate.Saved != 0 || populate.Checked != 0 {
			t.Fatalf("attempt %d: Populate counted work (checked %d, saved %d) although every shard RPC fails",
				attempt, populate.Checked, populate.Saved)
		}
	}
}
