package tasks

import (
	"net"
	"testing"

	db_server "github.com/memocash/index/db/server"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/test/suite"
)

// TestSuiteStartUnwindsPartialStartup occupies the second shard's port so
// suite startup fails after the first shard is already running, then verifies
// the partial startup was fully unwound: once the port is released, a fresh
// suite must start cleanly in the same process. This lives in the tasks
// package deliberately - suite-backed tests share fixed ports, and keeping
// them in one test binary keeps them sequential.
func TestSuiteStartUnwindsPartialStartup(t *testing.T) {
	t.Chdir(t.TempDir())
	shards := config.GetQueueShards()
	if len(shards) != 2 {
		t.Fatalf("expected 2 queue shards, got %d", len(shards))
	}
	blocker, err := net.Listen("tcp", db_server.GetListenHost(shards[1].Port))
	if err != nil {
		t.Fatalf("error occupying shard 1 port: %v", err)
	}
	s := suite.GetNewSuite()
	if err := s.Start(); err == nil {
		s.EndPrint()
		blocker.Close()
		t.Fatal("expected suite start to fail with shard 1 port occupied")
	}
	if err := blocker.Close(); err != nil {
		t.Fatalf("error releasing shard 1 port: %v", err)
	}
	fresh := suite.GetNewSuite()
	if err := fresh.Start(); err != nil {
		t.Fatalf("fresh suite failed to start after unwound partial startup: %v", err)
	}
	if err := fresh.End(); err != nil {
		t.Fatalf("error ending fresh suite: %v", err)
	}
}
