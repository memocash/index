package script

import (
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LinkRevoke struct {
	AcceptTxHash []byte
	Message      string
}

func (l LinkRevoke) Get() ([]byte, error) {
	if len(l.AcceptTxHash) != memo.TxHashLength {
		return nil, fmt.Errorf("incorrect accept tx hash size: %d", len(l.AcceptTxHash))
	}
	var msgByte = []byte(l.Message)
	var maxSize = memo.OldMaxReplySize
	if len(msgByte) > maxSize {
		return nil, fmt.Errorf("error message too big %d, max %d", len(msgByte), maxSize)
	}
	script := memo.GetBaseOpReturn().
		AddData(memo.PrefixLinkRevoke).
		AddData(jutil.ByteReverse(l.AcceptTxHash))
	if len(l.Message) > 0 {
		script = script.AddData(msgByte)
	}
	pkScript, err := script.Script()
	if err != nil {
		return nil, fmt.Errorf("error creating link revoke script; %w", err)
	}
	return pkScript, nil
}

func (l LinkRevoke) Type() memo.OutputType {
	return memo.OutputTypeLinkRevoke
}
