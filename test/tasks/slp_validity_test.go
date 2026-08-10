package tasks

import (
	"testing"

	"github.com/memocash/index/test/suite"
)

// TestSlpValidity runs the full SLP validity end-to-end task - real queue
// shard servers, the real saver pipeline, cascading validation, and the
// sweeper - as part of the standard test suite so none of it can regress
// silently. The same task is runnable interactively via the debug build
// (`go build -tags debug`, then `./index debug test slp_validity`).
func TestSlpValiditySuite(t *testing.T) {
	t.Chdir(t.TempDir()) // keep the suite's leveldb data out of the repo tree
	if err := suite.Run(&SlpValidity, nil); err != nil {
		t.Fatalf("error running slp validity suite; %v", err)
	}
}
