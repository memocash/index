package script

import (
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LinkAccept struct {
	RequestTxHash []byte
	Message       string
}

func (l LinkAccept) Get() ([]byte, error) {
	if len(l.RequestTxHash) != memo.TxHashLength {
		return nil, fmt.Errorf("incorrect request tx hash size: %d", len(l.RequestTxHash))
	}
	var msgByte = []byte(l.Message)
	var maxSize = memo.OldMaxReplySize
	if len(msgByte) > maxSize {
		return nil, fmt.Errorf("error message too big %d, max %d", len(msgByte), maxSize)
	}
	script := memo.GetBaseOpReturn().
		AddData(memo.PrefixLinkAccept).
		AddData(jutil.ByteReverse(l.RequestTxHash))
	if len(l.Message) > 0 {
		script = script.AddData(msgByte)
	}
	pkScript, err := script.Script()
	if err != nil {
		return nil, fmt.Errorf("error creating link accept script; %w", err)
	}
	return pkScript, nil
}

func (l LinkAccept) Type() memo.OutputType {
	return memo.OutputTypeLinkAccept
}
