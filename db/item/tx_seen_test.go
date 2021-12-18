package item_test

import (
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"testing"
	"time"
)

type TxSeenTest struct {
	Raw      string
	Expected time.Time
}

func TestTxSeen_Deserialize(t *testing.T) {
	var Raw1 = TxSeenTest{
		Raw:      "0051a1b9477745e1dec79aec96d3de7e9cf782668f5048fe20123902807c93464ff37e6100000000",
		Expected: time.Date(2021, 10, 31, 12, 49, 35, 0, time.Local),
	}
	var Raw2 = TxSeenTest{
		Raw:      "c25ff6da337fd681c7722023edfc6ba347780c8a6177fc638baddae4fbc94b15e0693b540bb0b816",
		Expected: time.Date(2021, 11, 18, 8, 13, 47, 253484000, time.Local),
	}
	for _, raw := range []TxSeenTest{Raw1, Raw2} {
		uid, _ := hex.DecodeString(raw.Raw)
		var txSeen = new(item.TxSeen)
		txSeen.SetUid(uid)
		if !txSeen.Timestamp.Equal(raw.Expected) {
			t.Error(jerr.Newf("error timestamp does not match expected: %s %s", txSeen.Timestamp, raw.Expected))
		}
	}
}
