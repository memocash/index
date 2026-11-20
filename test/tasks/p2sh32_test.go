package tasks

import (
	"testing"

	"github.com/memocash/index/test/suite"
)

// TestP2sh32Suite runs the P2SH32 end-to-end task against real queue shard
// servers. Also runnable via the debug build:
// `go build -tags debug`, then `./index debug test p2sh32`.
func TestP2sh32Suite(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := suite.Run(&P2sh32, nil); err != nil {
		t.Fatalf("error running p2sh32 suite; %v", err)
	}
}
