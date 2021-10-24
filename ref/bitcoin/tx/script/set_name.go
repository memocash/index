package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type SetName struct {
	Name string
}

func (p SetName) Get() ([]byte, error) {
	name := []byte(p.Name)
	if len(name) > memo.OldMaxPostSize {
		return nil, jerr.New("name too long")
	}
	if len(name) == 0 {
		return nil, jerr.New("empty name")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSetName).
		AddData(name).
		Script()
	if err != nil {
		return nil, jerr.Get("error building set name script", err)
	}
	return pkScript, nil
}

func (p SetName) Type() memo.OutputType {
	return memo.OutputTypeMemoSetName
}
