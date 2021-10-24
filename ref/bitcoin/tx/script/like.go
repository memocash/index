package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type Like struct {
	TxHash []byte
}

func (l Like) Get() ([]byte, error) {
	data := l.TxHash
	if len(data) > memo.MaxPostSize {
		return nil, jerr.New("data too large")
	}
	if len(data) == 0 {
		return nil, jerr.New("empty data")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixLike).
		AddData(data).
		Script()
	if err != nil {
		return nil, jerr.Get("error creating memo like script", err)
	}
	return pkScript, nil
}

func (l Like) Type() memo.OutputType {
	return memo.OutputTypeMemoLike
}
