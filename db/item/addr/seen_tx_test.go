package addr_test

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/server"
	"github.com/memocash/index/db/store"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"testing"
)

type TestServer struct {
	Server *server.Server
}

func (s *TestServer) Init(topic string, shard uint) error {
	testDbPath := filepath.Join(os.TempDir(), fmt.Sprintf("goleveldbtest-%d", os.Getuid()))
	if err := os.RemoveAll(testDbPath); err != nil {
		return fmt.Errorf("error removing old db; %w", err)
	}

	testDb, err := leveldb.OpenFile(testDbPath, nil)
	if err != nil {
		return fmt.Errorf("error opening level db; %w", err)
	}

	store.SetConn(store.GetConnId(topic, shard), testDb)

	s.Server = server.NewServer(0, shard)

	return nil
}

func (s *TestServer) Run(t *testing.T) {
	go func() {
		t.Error(s.Server.Run())
	}()
}

func TestSeenTx_GetTopic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ts := &TestServer{}
	if err := ts.Init(db.TopicAddrSeenTx, 0); err != nil {
		t.Error(err)
		return
	}
	defer ts.Server.Stop()
	// TODO: Implement
	seenTxs, err := addr.GetSeenTxs(ctx, [25]byte{}, []byte{})
	if err != nil {
		t.Error(err)
		return
	}
	if len(seenTxs) != 0 {
		t.Error("Expected 0 seen txs")
	}
}
