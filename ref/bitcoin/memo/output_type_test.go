package memo_test

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"log"
	"testing"
)

type OutputTypeTest struct {
	Script []byte
	Type   memo.OutputType
}

func (tst OutputTypeTest) Test(t *testing.T) {
	outputTypeNew := memo.GetOutputType(tst.Script)
	if outputTypeNew != tst.Type {
		t.Error(fmt.Errorf("OutputType new %s does not match expected %s", outputTypeNew, tst.Type))
	}
	if testing.Verbose() {
		log.Printf("OutputTypeNew %s, expected %s\n", outputTypeNew, tst.Type)
	}
	outputType := memo.GetOutputTypeNew(tst.Script)
	if outputType != tst.Type {
		t.Error(fmt.Errorf("OutputType %s does not match expected %s", outputType, tst.Type))
	}
	if testing.Verbose() {
		log.Printf("OutputType %s, expected %s\n", outputType, tst.Type)
	}
}

func TestOutputTypeMessage(t *testing.T) {
	scr, _ := script.Post{Message: test_tx.TestMessage}.Get()
	OutputTypeTest{
		Script: scr,
		Type:   memo.OutputTypeMemoMessage,
	}.Test(t)
}

func TestOutputTypeLike(t *testing.T) {
	scr, _ := script.Like{TxHash: test_tx.GenericTxHash0}.Get()
	OutputTypeTest{
		Script: scr,
		Type:   memo.OutputTypeMemoLike,
	}.Test(t)
}

func TestOutputTypeP2pkh(t *testing.T) {
	scr, _ := script.P2pkh{PkHash: test_tx.Address1pkHash}.Get()
	OutputTypeTest{
		Script: scr,
		Type:   memo.OutputTypeUnknown, // TODO: P2pkh not supported
	}.Test(t)
}
