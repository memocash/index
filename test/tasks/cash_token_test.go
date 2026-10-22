package tasks

import (
	"testing"

	"github.com/memocash/index/test/suite"
)

// TestCashTokenSuite runs the CashTokens end-to-end task against real queue
// shard servers. Also runnable via the debug build:
// `go build -tags debug`, then `./index debug test cash_token`.
func TestCashTokenSuite(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := suite.Run(&CashToken, nil); err != nil {
		t.Fatalf("error running cash token suite; %v", err)
	}
}
