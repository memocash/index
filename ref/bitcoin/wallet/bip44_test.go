package wallet_test

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_wallet"
	"github.com/memocash/server/ref/bitcoin/wallet"
	"testing"
)

type Bip44PathTest struct {
	CoinType uint
	Index    uint
	Change   bool
	Path     string
}

var bip44PathTests = []Bip44PathTest{{
	CoinType: wallet.Bip44CoinTypeBTC,
	Index:    0,
	Change:   false,
	Path:     test_wallet.BtcPathAddress0,
}, {
	CoinType: wallet.Bip44CoinTypeBTC,
	Index:    1,
	Change:   false,
	Path:     test_wallet.BtcPathAddress1,
}, {
	CoinType: wallet.Bip44CoinTypeBTC,
	Index:    0,
	Change:   true,
	Path:     test_wallet.BtcPathChange0,
}, {
	CoinType: wallet.Bip44CoinTypeSLP,
	Index:    0,
	Change:   false,
	Path:     test_wallet.SlpPathAddress0,
}}

func TestGetBip44Path(t *testing.T) {
	for _, bip44PathTest := range bip44PathTests {
		path := wallet.GetBip44Path(bip44PathTest.CoinType, bip44PathTest.Index, bip44PathTest.Change)
		if path != bip44PathTest.Path {
			t.Error(jerr.Newf("bip44 path %s does not match expected %s", path, bip44PathTest.Path))
		} else if testing.Verbose() {
			jlog.Logf("bip44 path %s, expected %s\n", path, bip44PathTest.Path)
		}
	}
}
