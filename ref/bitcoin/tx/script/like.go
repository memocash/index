package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Like struct {
	TxHash []byte
}

func (l Like) Get() ([]byte, error) {
	data := l.TxHash
	if len(data) > memo.MaxPostSize {
		return nil, fmt.Errorf("data too large")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixLike).
		AddData(data).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error creating memo like script; %w", err)
	}
	return pkScript, nil
}

func (l Like) Type() memo.OutputType {
	return memo.OutputTypeMemoLike
}
