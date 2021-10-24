package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type Send struct {
	Hash    []byte
	Message string
}

func (p Send) Get() ([]byte, error) {
	if len(p.Hash) != memo.PkHashLength || len(p.Hash) != memo.ScriptHashLength {
		return nil, jerr.Newf("pk hash incorrect length %d", len(p.Hash))
	}
	message := []byte(p.Message)
	if len(message) > memo.OldMaxSendSize {
		return nil, jerr.New("data too large")
	}
	if len(message) == 0 {
		return nil, jerr.New("empty message")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSendMoney).
		AddData(p.Hash).
		AddData(message).
		Script()
	if err != nil {
		return nil, jerr.Get("error building send script", err)
	}
	return pkScript, nil
}

func (p Send) Type() memo.OutputType {
	return memo.OutputTypeMemoSend
}
