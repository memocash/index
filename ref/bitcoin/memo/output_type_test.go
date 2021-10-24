package memo_test

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type OutputTypeTest struct {
	Script []byte
	Type   memo.OutputType
}

func (tst OutputTypeTest) Test(t *testing.T) {
	outputTypeNew := memo.GetOutputType(tst.Script)
	if outputTypeNew != tst.Type {
		t.Error(jerr.Newf("OutputType new %s does not match expected %s", outputTypeNew, tst.Type))
	}
	if testing.Verbose() {
		jlog.Logf("OutputTypeNew %s, expected %s\n", outputTypeNew, tst.Type)
	}
	outputType := memo.GetOutputTypeNew(tst.Script)
	if outputType != tst.Type {
		t.Error(jerr.Newf("OutputType %s does not match expected %s", outputType, tst.Type))
	}
	if testing.Verbose() {
		jlog.Logf("OutputType %s, expected %s\n", outputType, tst.Type)
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
