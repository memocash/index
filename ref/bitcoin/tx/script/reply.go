package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Reply struct {
	TxHash  []byte
	Message string
}

func (r Reply) Get() ([]byte, error) {
	if len(r.Message) > memo.MaxReplySize {
		return nil, jerr.New("reply message too large")
	}
	if len(r.Message) == 0 {
		return nil, jerr.New("empty message")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixReply).
		AddData(r.TxHash).
		AddData([]byte(r.Message)).
		Script()
	if err != nil {
		return nil, jerr.Get("error creating memo reply output", err)
	}
	return pkScript, nil
}

func (r Reply) Type() memo.OutputType {
	return memo.OutputTypeMemoReply
}
