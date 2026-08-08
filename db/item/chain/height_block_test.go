package chain

import (
	"bytes"
	"encoding/binary"
	"sort"
	"testing"

	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
)

func heightBlockTestUid(height int64, n uint32) []byte {
	var blockHash [32]byte
	binary.BigEndian.PutUint32(blockHash[:4], n)
	return jutil.CombineBytes(jutil.GetInt64DataBig(height), jutil.ByteReverse(blockHash[:]))
}

// fakeShardFetch simulates a single shard's store: sorted uids, inclusive
// start bound, at most limit records per page (a short page only when the
// iterator is exhausted - the contract paginateHeightBlocks relies on).
func fakeShardFetch(uids [][]byte) func(start []byte, limit int) ([]client.Message, error) {
	return func(start []byte, limit int) ([]client.Message, error) {
		var messages []client.Message
		for _, uid := range uids {
			if bytes.Compare(uid, start) < 0 {
				continue
			}
			messages = append(messages, client.Message{Uid: uid})
			if len(messages) == limit {
				break
			}
		}
		return messages, nil
	}
}

// TestPaginateHeightBlocksDuplicateHeavyHeight covers the finding that a
// height with many duplicate block records (all on one shard, since height
// blocks shard by height) can exceed any single page: pagination must advance
// by the full compound height/block uid and return every record exactly once,
// so no record is missed before its height is checkpointed.
func TestPaginateHeightBlocksDuplicateHeavyHeight(t *testing.T) {
	var uids [][]byte
	for height := int64(0); height < 5; height++ {
		uids = append(uids, heightBlockTestUid(height, uint32(height)))
	}
	// A single height with more records than two full pages
	var duplicateCount = client.HugeLimit*2 + 500
	for n := 0; n < duplicateCount; n++ {
		uids = append(uids, heightBlockTestUid(5, uint32(1000+n)))
	}
	for height := int64(6); height < 10; height++ {
		uids = append(uids, heightBlockTestUid(height, uint32(height)), heightBlockTestUid(height, uint32(height+100)))
	}
	// Store iterates in uid order
	sort.Slice(uids, func(i, j int) bool { return bytes.Compare(uids[i], uids[j]) < 0 })
	heightBlocks, err := paginateHeightBlocks(fakeShardFetch(uids), 0, 8)
	if err != nil {
		t.Fatalf("paginate: %v", err)
	}
	var expected = 5 + duplicateCount + 2*2 // heights 0-4, 5, and 6-7
	if len(heightBlocks) != expected {
		t.Fatalf("expected %d records, got %d", expected, len(heightBlocks))
	}
	var seen = make(map[[32]byte]bool)
	for _, heightBlock := range heightBlocks {
		if heightBlock.Height < 0 || heightBlock.Height >= 8 {
			t.Fatalf("record height %d outside requested range", heightBlock.Height)
		}
		if seen[heightBlock.BlockHash] {
			t.Fatalf("record returned twice at height %d", heightBlock.Height)
		}
		seen[heightBlock.BlockHash] = true
	}
}

// TestPaginateHeightBlocksBounds covers the start and end bounds: records
// below startHeight and at or above endHeight are excluded.
func TestPaginateHeightBlocksBounds(t *testing.T) {
	var uids [][]byte
	for height := int64(0); height < 10; height++ {
		uids = append(uids, heightBlockTestUid(height, uint32(height)))
	}
	heightBlocks, err := paginateHeightBlocks(fakeShardFetch(uids), 3, 7)
	if err != nil {
		t.Fatalf("paginate: %v", err)
	}
	if len(heightBlocks) != 4 {
		t.Fatalf("expected 4 records, got %d", len(heightBlocks))
	}
	for i, heightBlock := range heightBlocks {
		if heightBlock.Height != int64(3+i) {
			t.Fatalf("expected height %d at position %d, got %d", 3+i, i, heightBlock.Height)
		}
	}
}
