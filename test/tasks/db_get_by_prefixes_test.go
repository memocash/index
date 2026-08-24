package tasks

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/test/run/queue"
	"github.com/memocash/index/test/suite"
)

func saveQueueItems(t *testing.T, shard uint32, itemTopic string, uids ...string) {
	t.Helper()
	var items = make([]queue.Item, len(uids))
	for i, uid := range uids {
		items[i] = queue.Item{Topic: itemTopic, Uid: []byte(uid), Data: []byte("data-" + uid)}
	}
	if err := queue.NewAdd(shard).Add(items); err != nil {
		t.Fatalf("error saving queue items; %v", err)
	}
}

// legacyByPrefixes queries one shard through the deprecated prefix RPC, the
// behavior baseline db.GetByPrefixes' GetFiltered translation must match.
func legacyByPrefixes(t *testing.T, shard uint32, itemTopic string, prefixes []client.Prefix,
	opts ...client.Option) []client.Message {
	t.Helper()
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	c := client.NewClient(shardConfig.GetHost())
	if err := c.GetByPrefixes(context.Background(), itemTopic, prefixes, opts...); err != nil {
		t.Fatalf("error getting legacy prefix messages; %v", err)
	}
	return c.Messages
}

func dbByPrefixes(t *testing.T, itemTopic string, shardPrefixes map[uint32][]client.Prefix,
	opts ...client.Option) []client.Message {
	t.Helper()
	messages, err := db.GetByPrefixes(context.Background(), itemTopic, shardPrefixes, opts...)
	if err != nil {
		t.Fatalf("error getting db prefix messages; %v", err)
	}
	return messages
}

func compareMessageLists(t *testing.T, got, want []client.Message) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("message count = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if !bytes.Equal(got[i].Uid, want[i].Uid) || !bytes.Equal(got[i].Message, want[i].Message) {
			t.Fatalf("message %d = %s/%s, want %s/%s",
				i, got[i].Uid, got[i].Message, want[i].Uid, want[i].Message)
		}
	}
}

func sortMessagesByUid(messages []client.Message) {
	sort.Slice(messages, func(i, j int) bool {
		return bytes.Compare(messages[i].Uid, messages[j].Uid) < 0
	})
}

// TestDbGetByPrefixes pins db.GetByPrefixes' translation onto the GetFiltered
// RPC against real shard servers by comparing it with the legacy prefix RPC:
// shard fan-out, resume cursors, per-prefix and request-wide limits, order,
// the legacy HugeLimit default cap, batches above HugeLimit prefixes, and the
// one documented divergence (an out-of-prefix Start is a bound, not ignored).
// It lives in the tasks package because suite-backed tests share fixed ports
// and must stay sequential in one test binary.
func TestDbGetByPrefixes(t *testing.T) {
	t.Chdir(t.TempDir())
	s := suite.GetNewSuite()
	if err := s.Start(); err != nil {
		t.Fatalf("error starting suite; %v", err)
	}
	defer s.EndPrint()

	const basicTopic = "prefix-basic"
	var shard0Uids = []string{
		"aa-0", "aa-1", "aa-2", "aa-3", "aa-4", "aa-5", "aa-6", "aa-7", "aa-8", "aa-9",
		"ab-0", "ab-1", "ab-2", "ab-3", "ab-4",
	}
	saveQueueItems(t, 0, basicTopic, shard0Uids...)
	saveQueueItems(t, 1, basicTopic, "ba-0", "ba-1", "ba-2", "ba-3", "ba-4", "ba-5", "ba-6")

	t.Run("multi shard fan out", func(t *testing.T) {
		got := dbByPrefixes(t, basicTopic, map[uint32][]client.Prefix{
			0: {client.NewPrefix([]byte("aa")), client.NewPrefix([]byte("ab"))},
			1: {client.NewPrefix([]byte("ba"))},
		})
		legacy := append(
			legacyByPrefixes(t, 0, basicTopic,
				[]client.Prefix{client.NewPrefix([]byte("aa")), client.NewPrefix([]byte("ab"))}),
			legacyByPrefixes(t, 1, basicTopic, []client.Prefix{client.NewPrefix([]byte("ba"))})...)
		// Cross-shard append order depends on goroutine completion; compare
		// as sets
		sortMessagesByUid(got)
		sortMessagesByUid(legacy)
		compareMessageLists(t, got, legacy)
		if len(got) != 22 {
			t.Errorf("got %d messages, want 22", len(got))
		}
	})

	t.Run("resume cursor extending prefix", func(t *testing.T) {
		prefixes := []client.Prefix{{Prefix: []byte("aa"), Start: []byte("aa-3\x00")}}
		got := dbByPrefixes(t, basicTopic, map[uint32][]client.Prefix{0: prefixes})
		compareMessageLists(t, got, legacyByPrefixes(t, 0, basicTopic, prefixes))
		if len(got) != 6 || !bytes.Equal(got[0].Uid, []byte("aa-4")) {
			t.Errorf("cursor resume returned %d messages, want 6 starting at aa-4", len(got))
		}
	})

	t.Run("per prefix limit", func(t *testing.T) {
		prefixes := []client.Prefix{
			{Prefix: []byte("aa"), Limit: 3},
			{Prefix: []byte("ab"), Limit: 2},
		}
		got := dbByPrefixes(t, basicTopic, map[uint32][]client.Prefix{0: prefixes})
		compareMessageLists(t, got, legacyByPrefixes(t, 0, basicTopic, prefixes))
		if len(got) != 5 {
			t.Errorf("got %d messages, want 5", len(got))
		}
	})

	t.Run("request limit and desc order options", func(t *testing.T) {
		prefixes := []client.Prefix{client.NewPrefix([]byte("aa")), client.NewPrefix([]byte("ab"))}
		gotLimit := dbByPrefixes(t, basicTopic, map[uint32][]client.Prefix{0: prefixes},
			client.NewOptionLimit(4))
		compareMessageLists(t, gotLimit, legacyByPrefixes(t, 0, basicTopic, prefixes,
			client.NewOptionLimit(4)))
		if len(gotLimit) != 4 {
			t.Errorf("got %d messages with request limit, want 4", len(gotLimit))
		}
		gotDesc := dbByPrefixes(t, basicTopic, map[uint32][]client.Prefix{0: prefixes},
			client.NewOptionOrder(true))
		compareMessageLists(t, gotDesc, legacyByPrefixes(t, 0, basicTopic, prefixes,
			client.NewOptionOrder(true)))
		if len(gotDesc) != 15 || !bytes.Equal(gotDesc[0].Uid, []byte("aa-9")) {
			t.Errorf("desc returned %d messages first %s, want 15 starting at aa-9",
				len(gotDesc), gotDesc[0].Uid)
		}
	})

	// The one documented divergence from the legacy path: a Start not
	// extending its prefix used to be silently ignored, through GetFiltered
	// it is a real bound. No production caller passes one (resume cursors are
	// built from returned uids), so the old behavior is pinned here only as
	// documentation of the difference.
	t.Run("out of prefix start is a bound not ignored", func(t *testing.T) {
		prefixes := []client.Prefix{{Prefix: []byte("aa"), Start: []byte("zz")}}
		got := dbByPrefixes(t, basicTopic, map[uint32][]client.Prefix{0: prefixes})
		if len(got) != 0 {
			t.Errorf("got %d messages with out-of-prefix start, want 0 (start is a bound)", len(got))
		}
		legacy := legacyByPrefixes(t, 0, basicTopic, prefixes)
		if len(legacy) != 10 {
			t.Errorf("legacy got %d messages, want 10 (start ignored)", len(legacy))
		}
	})

	t.Run("absent per prefix limit keeps legacy HugeLimit cap", func(t *testing.T) {
		const capTopic = "prefix-cap"
		var uids = make([]string, client.HugeLimit+10)
		for i := range uids {
			uids[i] = fmt.Sprintf("big-%05d", i)
		}
		saveQueueItems(t, 0, capTopic, uids...)
		prefixes := []client.Prefix{client.NewPrefix([]byte("big-"))}
		got := dbByPrefixes(t, capTopic, map[uint32][]client.Prefix{0: prefixes})
		if len(got) != client.HugeLimit {
			t.Errorf("got %d messages with no per-prefix limit, want the %d cap",
				len(got), client.HugeLimit)
		}
		compareMessageLists(t, got, legacyByPrefixes(t, 0, capTopic, prefixes))
	})

	t.Run("prefix batches above HugeLimit", func(t *testing.T) {
		const batchTopic = "prefix-batch"
		var count = client.HugeLimit + 1
		var uids = make([]string, count)
		var prefixes = make([]client.Prefix, count)
		for i := range uids {
			uids[i] = fmt.Sprintf("p%05d-x", i)
			prefixes[i] = client.NewPrefix([]byte(fmt.Sprintf("p%05d-", i)))
		}
		saveQueueItems(t, 0, batchTopic, uids...)
		got := dbByPrefixes(t, batchTopic, map[uint32][]client.Prefix{0: prefixes})
		if len(got) != count {
			t.Fatalf("got %d messages across prefix batches, want %d", len(got), count)
		}
		// A shard's batches run sequentially, so arm order holds across the
		// batch boundary
		for i := range got {
			if string(got[i].Uid) != uids[i] {
				t.Fatalf("message %d uid = %s, want %s", i, got[i].Uid, uids[i])
			}
		}
	})
}
