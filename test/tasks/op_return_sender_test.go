package tasks

import (
	"testing"

	"github.com/memocash/index/test/suite"
)

// TestOpReturnSenderSuite runs the memo sender attribution task - real queue
// shard servers and the real saver pipeline - so first-input sender selection
// and the no-parseable-input skip can't regress silently. Also runnable
// interactively via the debug build (`go build -tags debug`, then
// `./index debug test op_return_sender`).
func TestOpReturnSenderSuite(t *testing.T) {
	t.Chdir(t.TempDir()) // keep the suite's leveldb data out of the repo tree
	if err := suite.Run(&OpReturnSender, nil); err != nil {
		t.Fatalf("error running op return sender suite; %v", err)
	}
}
