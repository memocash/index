package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Profile struct {
	Text string
}

func (p Profile) Get() ([]byte, error) {
	text := []byte(p.Text)
	if len(text) > memo.OldMaxPostSize {
		return nil, jerr.New("text size too large")
	}
	if len(text) == 0 {
		return nil, jerr.New("empty text")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSetProfile).
		AddData(text).
		Script()
	if err != nil {
		return nil, jerr.Get("error building script", err)
	}
	return pkScript, nil
}

func (p Profile) Type() memo.OutputType {
	return memo.OutputTypeMemoSetProfile
}
