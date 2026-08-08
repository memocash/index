package maint

import "testing"

func hash(b byte) [32]byte {
	return [32]byte{b}
}

// TestSweepCursorCheckpointsOnlyCompletedHeights covers the crash-resume
// contract: a height is reported completed only once a higher height is
// reached, so a crash while a height's records are mid-flight resumes at that
// height (checkpoint+1) and re-processes all of its records.
func TestSweepCursorCheckpointsOnlyCompletedHeights(t *testing.T) {
	cursor := newSweepCursor()
	if _, ok := cursor.advance(100); ok {
		t.Fatal("first height must not report a completed predecessor")
	}
	if cursor.seen(hash(1)) {
		t.Fatal("first block at height must not read as already processed")
	}
	// Second record at the same height (duplicate/fork block): no checkpoint
	if _, ok := cursor.advance(100); ok {
		t.Fatal("same height must not complete itself")
	}
	if cursor.seen(hash(2)) {
		t.Fatal("second distinct block at height must not read as processed")
	}
	// Moving on completes 100 exactly once
	completed, ok := cursor.advance(101)
	if !ok || completed != 100 {
		t.Fatalf("expected height 100 completed, got %d %v", completed, ok)
	}
	if _, ok := cursor.advance(101); ok {
		t.Fatal("re-advancing to current height must not re-complete")
	}
	// Heights never move backward
	if _, ok := cursor.advance(99); ok {
		t.Fatal("moving backward must not complete anything")
	}
	if cursor.current != 101 {
		t.Fatalf("backward advance must not change current, got %d", cursor.current)
	}
	completed, ok = cursor.finish()
	if !ok || completed != 101 {
		t.Fatalf("expected finish to complete height 101, got %d %v", completed, ok)
	}
}

// TestSweepCursorPageBoundaryRefetch covers the inclusive page-boundary
// refetch: records of the in-progress height re-listed on the next page are
// deduplicated, while new records at that height still process.
func TestSweepCursorPageBoundaryRefetch(t *testing.T) {
	cursor := newSweepCursor()
	cursor.advance(200)
	if cursor.seen(hash(1)) {
		t.Fatal("first sighting must process")
	}
	// Next page re-lists height 200 (inclusive start): same block dedupes
	cursor.advance(200)
	if !cursor.seen(hash(1)) {
		t.Fatal("re-listed block must be deduplicated")
	}
	if cursor.seen(hash(2)) {
		t.Fatal("a new block at the in-progress height must still process")
	}
	// Dedup state resets when the height completes
	if completed, ok := cursor.advance(201); !ok || completed != 200 {
		t.Fatalf("expected height 200 completed, got %d %v", completed, ok)
	}
	if cursor.seen(hash(1)) {
		t.Fatal("dedup state must reset for a new height")
	}
}

// TestSweepCursorEmptyRun covers a sweep that finds nothing to process.
func TestSweepCursorEmptyRun(t *testing.T) {
	cursor := newSweepCursor()
	if _, ok := cursor.finish(); ok {
		t.Fatal("finish with no processed heights must not checkpoint")
	}
}
