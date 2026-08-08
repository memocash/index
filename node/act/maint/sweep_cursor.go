package maint

// sweepCursor tracks fully-completed heights for the slp validity sweep.
// Height-block records arrive in ascending height order but a height can span
// multiple records (duplicate/fork blocks), so a height may only be
// checkpointed once a higher height is reached (or the sweep ends). Resuming
// at checkpoint+1 then re-processes any partially-completed height instead of
// skipping it. The seen-dedup also guards against any record being listed
// twice across fetches.
type sweepCursor struct {
	current      int64 // height currently in progress; -1 = none yet
	doneAtHeight map[[32]byte]bool
}

func newSweepCursor() *sweepCursor {
	return &sweepCursor{
		current:      -1,
		doneAtHeight: make(map[[32]byte]bool),
	}
}

// advance moves to the given height and reports (completedHeight, true) when
// doing so finishes a previous height. Heights never move backward.
func (c *sweepCursor) advance(height int64) (int64, bool) {
	if height <= c.current {
		return 0, false
	}
	var prev = c.current
	c.current = height
	c.doneAtHeight = make(map[[32]byte]bool)
	if prev >= 0 {
		return prev, true
	}
	return 0, false
}

// seen marks a block at the current height as processed and reports whether it
// had already been processed (page-boundary re-fetches are inclusive).
func (c *sweepCursor) seen(blockHash [32]byte) bool {
	if c.doneAtHeight[blockHash] {
		return true
	}
	c.doneAtHeight[blockHash] = true
	return false
}

// finish reports the final in-progress height as completed, if any. Only call
// once no more records remain at or above it.
func (c *sweepCursor) finish() (int64, bool) {
	if c.current < 0 {
		return 0, false
	}
	return c.current, true
}
