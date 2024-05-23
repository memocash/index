package memo_test

import (
	"fmt"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"testing"
)

var slpQuantityTests = []struct {
	Quantity    uint64
	Decimals    uint8
	SlpQuantity string
}{{
	Quantity:    10,
	Decimals:    0,
	SlpQuantity: "10",
}, {
	Quantity:    1,
	Decimals:    8,
	SlpQuantity: "0.00000001",
}}

func TestGetSlpQuantity(t *testing.T) {
	for _, slpQuantityTest := range slpQuantityTests {
		slpQuantity := memo.GetSlpQuantity(slpQuantityTest.Quantity, slpQuantityTest.Decimals)
		var str = jfmt.AddCommasFloat(slpQuantity)
		if str != slpQuantityTest.SlpQuantity {
			t.Error(fmt.Errorf("slpQuantity (%s) does not match slpQuantityTest.SlpQuantity (%s)", str, slpQuantityTest.SlpQuantity))
		}
	}
}
