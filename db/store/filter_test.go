package store_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/memocash/index/db/store"
)

// Pattern constructors live in db/client (the store package never builds
// patterns itself); the tests build the same shapes locally.
func patternPrefix(b []byte) *store.Pattern {
	return &store.Pattern{Parts: [][]byte{b}, AnchorStart: true}
}

func patternContains(b []byte) *store.Pattern {
	return &store.Pattern{Parts: [][]byte{b}}
}

func patternSuffix(b []byte) *store.Pattern {
	return &store.Pattern{Parts: [][]byte{b}, AnchorEnd: true}
}

func TestPatternMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern *store.Pattern
		input   []byte
		want    bool
	}{
		{"nil pattern matches", nil, []byte("anything"), true},
		{"empty parts matches", &store.Pattern{}, []byte("anything"), true},
		{"empty parts matches empty", &store.Pattern{}, nil, true},
		{"contains hit", patternContains([]byte("ell")), []byte("hello"), true},
		{"contains miss", patternContains([]byte("elle")), []byte("hello"), false},
		{"contains percent byte", patternContains([]byte{0x25}), []byte{0x01, 0x25, 0x02}, true},
		{"contains percent byte miss", patternContains([]byte{0x25}), []byte{0x01, 0x02}, false},
		{"prefix hit", patternPrefix([]byte("he")), []byte("hello"), true},
		{"prefix miss mid-match", patternPrefix([]byte("el")), []byte("hello"), false},
		{"suffix hit", patternSuffix([]byte("lo")), []byte("hello"), true},
		{"suffix miss", patternSuffix([]byte("ll")), []byte("hello"), false},
		{"exact hit", &store.Pattern{Parts: [][]byte{[]byte("a")}, AnchorStart: true, AnchorEnd: true}, []byte("a"), true},
		{"exact miss repeat", &store.Pattern{Parts: [][]byte{[]byte("a")}, AnchorStart: true, AnchorEnd: true}, []byte("aa"), false},
		{"two parts gap", &store.Pattern{Parts: [][]byte{[]byte("ab"), []byte("cd")}}, []byte("xabycdz"), true},
		{"two parts out of order", &store.Pattern{Parts: [][]byte{[]byte("cd"), []byte("ab")}}, []byte("xabycdz"), false},
		{"anchored both A%B hit", &store.Pattern{Parts: [][]byte{[]byte("ab"), []byte("cd")}, AnchorStart: true, AnchorEnd: true}, []byte("ab-x-cd"), true},
		{"anchored both no gap", &store.Pattern{Parts: [][]byte{[]byte("ab"), []byte("cd")}, AnchorStart: true, AnchorEnd: true}, []byte("abcd"), true},
		{"suffix may not overlap prior part", &store.Pattern{Parts: [][]byte{[]byte("ab"), []byte("bc")}, AnchorStart: true, AnchorEnd: true}, []byte("abc"), false},
		{"suffix overlap of contains part", &store.Pattern{Parts: [][]byte{[]byte("a"), []byte("a")}, AnchorEnd: true}, []byte("aa"), true},
		{"suffix overlap of contains part miss", &store.Pattern{Parts: [][]byte{[]byte("a"), []byte("a")}, AnchorEnd: true}, []byte("a"), false},
		{"binary multi-part", &store.Pattern{Parts: [][]byte{{0x00, 0x25}, {0x25, 0xff}}}, []byte{0x00, 0x25, 0x01, 0x25, 0xff}, true},
		{"empty part matches trivially", &store.Pattern{Parts: [][]byte{{}}}, []byte("x"), true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.pattern.Match(test.input); got != test.want {
				t.Errorf("Match(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}

var binaryMessages = []*store.Message{
	{Uid: []byte{0x10, 0x25, 0x01}, Message: []byte{0xaa, 0x25, 0xbb}},
	{Uid: []byte{0x10, 0x25, 0x02}, Message: []byte{0xaa, 0xbb}},
	{Uid: []byte{0x11, 0x01}, Message: []byte{0x25}},
}

func initFilterTestDb(t *testing.T) {
	if err := initTestDb(); err != nil {
		t.Fatalf("error initializing test db; %v", err)
	}
	if err := store.SaveMessages(TestTopic, TestShard, binaryMessages); err != nil {
		t.Fatalf("error saving binary messages; %v", err)
	}
}

func getFiltered(t *testing.T, request store.FilterRequest) []*store.Message {
	t.Helper()
	request.Topic = TestTopic
	request.Shard = TestShard
	messages, err := store.GetFiltered(context.Background(), request)
	if err != nil {
		t.Fatalf("error getting filtered messages; %v", err)
	}
	return messages
}

func getFilteredSingle(t *testing.T, pattern store.FilterPattern, limit int) []*store.Message {
	t.Helper()
	return getFiltered(t, store.FilterRequest{Patterns: []store.FilterPattern{pattern}, Limit: limit})
}

func checkUids(t *testing.T, messages []*store.Message, expected ...[]byte) {
	t.Helper()
	if len(messages) != len(expected) {
		t.Fatalf("message count = %d, want %d", len(messages), len(expected))
	}
	for i := range messages {
		if !bytes.Equal(messages[i].Uid, expected[i]) {
			t.Errorf("message %d uid = %x, want %x", i, messages[i].Uid, expected[i])
		}
	}
}

// nextStartAsc is the ascending keyset cursor: the smallest uid strictly
// greater than the last returned one.
func nextStartAsc(messages []*store.Message) []byte {
	last := messages[len(messages)-1]
	return append(store.GetPtrSlice(last.Uid), 0x00)
}

func TestGetFiltered(t *testing.T) {
	initFilterTestDb(t)
	defer store.CloseAll()

	t.Run("uid suffix", func(t *testing.T) {
		messages := getFilteredSingle(t, store.FilterPattern{Uid: patternSuffix([]byte("-3"))}, 0)
		checkUids(t, messages, testMessageOther3.Uid, testMessageTest3.Uid)
	})

	t.Run("uid contains binary percent", func(t *testing.T) {
		messages := getFilteredSingle(t, store.FilterPattern{Uid: patternContains([]byte{0x25})}, 0)
		checkUids(t, messages, binaryMessages[0].Uid, binaryMessages[1].Uid)
	})

	t.Run("data filter only", func(t *testing.T) {
		messages := getFilteredSingle(t, store.FilterPattern{Data: patternContains([]byte("message-7"))}, 0)
		checkUids(t, messages, testMessageOther7.Uid, testMessageTest7.Uid)
	})

	t.Run("anchored uid with data filter", func(t *testing.T) {
		messages := getFilteredSingle(t, store.FilterPattern{
			Uid:  patternPrefix([]byte(PrefixTest)),
			Data: patternSuffix([]byte("-5")),
		}, 0)
		checkUids(t, messages, testMessageTest5.Uid)
	})

	t.Run("prefix parity with deprecated GetByPrefixes", func(t *testing.T) {
		legacy, err := store.GetByPrefixes(store.RequestByPrefixes{
			Topic:    TestTopic,
			Shard:    TestShard,
			Prefixes: []store.Prefix{{Prefix: []byte(PrefixTest)}},
		})
		if err != nil {
			t.Fatalf("error getting legacy prefix messages; %v", err)
		}
		pattern := getFilteredSingle(t, store.FilterPattern{Uid: patternPrefix([]byte(PrefixTest))}, 0)
		if len(legacy) != len(pattern) {
			t.Fatalf("legacy count = %d, pattern count = %d", len(legacy), len(pattern))
		}
		for i := range legacy {
			if !bytes.Equal(legacy[i].Uid, pattern[i].Uid) {
				t.Errorf("message %d uid mismatch: %x vs %x", i, legacy[i].Uid, pattern[i].Uid)
			}
		}
	})

	t.Run("multiple arms return in arm order", func(t *testing.T) {
		messages := getFiltered(t, store.FilterRequest{Patterns: []store.FilterPattern{
			{Uid: patternSuffix([]byte("-3"))},
			{Uid: patternSuffix([]byte("-7"))},
		}})
		checkUids(t, messages,
			testMessageOther3.Uid, testMessageTest3.Uid,
			testMessageOther7.Uid, testMessageTest7.Uid)
	})

	t.Run("request limit truncates across arms", func(t *testing.T) {
		messages := getFiltered(t, store.FilterRequest{Patterns: []store.FilterPattern{
			{Uid: patternSuffix([]byte("-3"))},
			{Uid: patternSuffix([]byte("-7"))},
		}, Limit: 3})
		checkUids(t, messages,
			testMessageOther3.Uid, testMessageTest3.Uid, testMessageOther7.Uid)
	})

	t.Run("empty request matches all", func(t *testing.T) {
		messages := getFiltered(t, store.FilterRequest{})
		if len(messages) != 23 {
			t.Errorf("got %d messages, want 23", len(messages))
		}
	})

	// A part-less pattern is documented to match everything, and protobuf
	// accepts one with any anchor combination; the scanner must scan the
	// full range rather than panic indexing an absent first part
	t.Run("part-less uid pattern matches all for every anchor combination", func(t *testing.T) {
		for _, pattern := range []*store.Pattern{
			{},
			{AnchorStart: true},
			{AnchorEnd: true},
			{AnchorStart: true, AnchorEnd: true},
			{Parts: [][]byte{{}}, AnchorStart: true}, // empty first part
		} {
			messages := getFilteredSingle(t, store.FilterPattern{Uid: pattern}, 0)
			if len(messages) != 23 {
				t.Errorf("pattern %+v: got %d messages, want 23", pattern, len(messages))
			}
		}
	})

	t.Run("part-less data pattern matches all", func(t *testing.T) {
		messages := getFilteredSingle(t, store.FilterPattern{Data: &store.Pattern{AnchorStart: true}}, 0)
		if len(messages) != 23 {
			t.Errorf("got %d messages, want 23", len(messages))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		messages := getFilteredSingle(t, store.FilterPattern{Uid: patternContains([]byte("missing"))}, 0)
		if len(messages) != 0 {
			t.Errorf("got %d messages, want 0", len(messages))
		}
	})

	// Unlike the legacy prefix path (which silently ignores a Start that does
	// not extend its prefix — pinned in TestGetByPrefixesPaging's foreign
	// cursor cases), a filter arm's Start is always a plain bound intersected
	// with the anchored prefix range
	t.Run("start outside prefix acts as plain bound", func(t *testing.T) {
		// "test" sorts after every "other-*" uid: as an asc lower bound it
		// empties the range, as a desc upper bound it excludes nothing
		messages := getFilteredSingle(t, store.FilterPattern{
			Uid:   patternPrefix([]byte(PrefixOther)),
			Start: []byte(PrefixTest),
		}, 0)
		checkUids(t, messages)

		// "other" sorts before every "test-*" uid: as an asc lower bound it
		// excludes nothing
		messages = getFilteredSingle(t, store.FilterPattern{
			Uid:   patternPrefix([]byte(PrefixTest)),
			Start: []byte(PrefixOther),
		}, 0)
		if len(messages) != 10 {
			t.Errorf("below-prefix asc start: got %d messages, want 10", len(messages))
		}

		// desc: Start is the exclusive upper bound; below the prefix range it
		// empties it, above the prefix range it excludes nothing
		below, err := store.GetFiltered(context.Background(), store.FilterRequest{
			Topic:    TestTopic,
			Shard:    TestShard,
			Patterns: []store.FilterPattern{{Uid: patternPrefix([]byte(PrefixTest)), Start: []byte(PrefixOther)}},
			Desc:     true,
		})
		if err != nil {
			t.Fatalf("error getting desc below-prefix start; %v", err)
		}
		checkUids(t, below)
		above, err := store.GetFiltered(context.Background(), store.FilterRequest{
			Topic:    TestTopic,
			Shard:    TestShard,
			Patterns: []store.FilterPattern{{Uid: patternPrefix([]byte(PrefixOther)), Start: []byte(PrefixTest)}},
			Desc:     true,
		})
		if err != nil {
			t.Fatalf("error getting desc above-prefix start; %v", err)
		}
		if len(above) != 10 {
			t.Errorf("above-prefix desc start: got %d messages, want 10", len(above))
		}
	})

	t.Run("desc reverses order", func(t *testing.T) {
		messages, err := store.GetFiltered(context.Background(), store.FilterRequest{
			Topic:    TestTopic,
			Shard:    TestShard,
			Patterns: []store.FilterPattern{{Uid: patternSuffix([]byte("-3"))}},
			Desc:     true,
		})
		if err != nil {
			t.Fatalf("error getting filtered desc; %v", err)
		}
		checkUids(t, messages, testMessageTest3.Uid, testMessageOther3.Uid)
	})
}

// TestGetFilteredCanceled pins that a filter scan observes context
// cancellation: a sparse pattern iterates a whole topic, and an abandoned RPC
// must stop the server-side work instead of burning disk and CPU to the end
// of the range. The check runs on the first scanned row, so a dead context
// fails immediately regardless of topic size.
func TestGetFilteredCanceled(t *testing.T) {
	initFilterTestDb(t)
	defer store.CloseAll()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := store.GetFiltered(ctx, store.FilterRequest{
		Topic:    TestTopic,
		Shard:    TestShard,
		Patterns: []store.FilterPattern{{Uid: patternContains([]byte("missing"))}},
	})
	if err == nil {
		t.Fatalf("expected error for canceled filter scan, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestGetFilteredPaging pins the keyset-pagination contract in both
// directions: page with a limit, resume asc from last-returned-uid + 0x00 or
// desc from last-returned-uid (exclusive upper bound), stop on a short page;
// the union of pages equals one unpaged scan.
func TestGetFilteredPaging(t *testing.T) {
	initFilterTestDb(t)
	defer store.CloseAll()

	for _, desc := range []bool{false, true} {
		full, err := store.GetFiltered(context.Background(), store.FilterRequest{
			Topic:    TestTopic,
			Shard:    TestShard,
			Patterns: []store.FilterPattern{{Data: patternContains([]byte("message"))}},
			Desc:     desc,
		})
		if err != nil {
			t.Fatalf("error getting full scan; %v", err)
		}
		if len(full) != 20 {
			t.Fatalf("full scan count = %d, want 20", len(full))
		}

		for _, limit := range []int{2, 3, 7} {
			t.Run(fmt.Sprintf("desc %v limit %d", desc, limit), func(t *testing.T) {
				var union []*store.Message
				var start []byte
				for calls := 0; ; calls++ {
					if calls > len(full)+1 {
						t.Fatalf("paging loop did not terminate")
					}
					messages, err := store.GetFiltered(context.Background(), store.FilterRequest{
						Topic: TestTopic,
						Shard: TestShard,
						Patterns: []store.FilterPattern{{
							Start: start,
							Data:  patternContains([]byte("message")),
						}},
						Limit: limit,
						Desc:  desc,
					})
					if err != nil {
						t.Fatalf("error getting page; %v", err)
					}
					union = append(union, messages...)
					if len(messages) < limit {
						break // short page: range exhausted
					}
					if desc {
						start = messages[len(messages)-1].Uid
					} else {
						start = nextStartAsc(messages)
					}
				}
				if len(union) != len(full) {
					t.Fatalf("union count = %d, want %d", len(union), len(full))
				}
				for i := range union {
					if !bytes.Equal(union[i].Uid, full[i].Uid) {
						t.Errorf("union message %d uid = %x, want %x", i, union[i].Uid, full[i].Uid)
					}
				}
			})
		}
	}
}

func TestGetFilteredPatternMax(t *testing.T) {
	initFilterTestDb(t)
	defer store.CloseAll()

	messages := getFilteredSingle(t, store.FilterPattern{
		Data: patternContains([]byte("message")),
		Max:  5,
	}, 0)
	if len(messages) != 5 {
		t.Fatalf("message count = %d, want 5 (per-arm max)", len(messages))
	}
	rest := getFilteredSingle(t, store.FilterPattern{
		Start: nextStartAsc(messages),
		Data:  patternContains([]byte("message")),
	}, 0)
	if len(rest) != 15 {
		t.Errorf("rest count = %d, want 15", len(rest))
	}
}
