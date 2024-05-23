package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Post struct {
	Message string
}

func (p Post) Get() ([]byte, error) {
	message := []byte(p.Message)
	if len(message) > memo.MaxPostSize {
		return nil, fmt.Errorf("message size too large")
	}
	if len(message) == 0 {
		return nil, fmt.Errorf("empty message")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixPost).
		AddData(message).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building script; %w", err)
	}
	return pkScript, nil
}

func (p Post) Type() memo.OutputType {
	return memo.OutputTypeMemoMessage
}
