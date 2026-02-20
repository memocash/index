package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Send struct {
	Hash    []byte
	Message string
}

func (p Send) Get() ([]byte, error) {
	if len(p.Hash) != memo.PkHashLength && len(p.Hash) != memo.ScriptHashLength {
		return nil, fmt.Errorf("pk hash incorrect length %d", len(p.Hash))
	}
	message := []byte(p.Message)
	if len(message) > memo.OldMaxSendSize {
		return nil, fmt.Errorf("data too large")
	}
	if len(message) == 0 {
		return nil, fmt.Errorf("empty message")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSendMoney).
		AddData(p.Hash).
		AddData(message).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building send script; %w", err)
	}
	return pkScript, nil
}

func (p Send) Type() memo.OutputType {
	return memo.OutputTypeMemoSend
}
