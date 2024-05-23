package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Profile struct {
	Text string
}

func (p Profile) Get() ([]byte, error) {
	text := []byte(p.Text)
	if len(text) > memo.OldMaxPostSize {
		return nil, fmt.Errorf("text size too large")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSetProfile).
		AddData(text).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building script; %w", err)
	}
	return pkScript, nil
}

func (p Profile) Type() memo.OutputType {
	return memo.OutputTypeMemoSetProfile
}
