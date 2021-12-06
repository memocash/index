package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Post struct {
	Message string
}

func (p Post) Get() ([]byte, error) {
	message := []byte(p.Message)
	if len(message) > memo.MaxPostSize {
		return nil, jerr.New("message size too large")
	}
	if len(message) == 0 {
		return nil, jerr.New("empty message")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixPost).
		AddData(message).
		Script()
	if err != nil {
		return nil, jerr.Get("error building script", err)
	}
	return pkScript, nil
}

func (p Post) Type() memo.OutputType {
	return memo.OutputTypeMemoMessage
}
