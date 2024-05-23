package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type SetName struct {
	Name string
}

func (p SetName) Get() ([]byte, error) {
	name := []byte(p.Name)
	if len(name) > memo.OldMaxPostSize {
		return nil, fmt.Errorf("name too long")
	}
	if len(name) == 0 {
		return nil, fmt.Errorf("empty name")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSetName).
		AddData(name).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building set name script; %w", err)
	}
	return pkScript, nil
}

func (p SetName) Type() memo.OutputType {
	return memo.OutputTypeMemoSetName
}
