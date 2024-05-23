package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type TokenSell struct {
	InOuts []InOut
}

func (t TokenSell) Get() ([]byte, error) {
	script := memo.GetBaseOpReturn().
		AddData(memo.PrefixSellTokenMake)
	for _, inOut := range t.InOuts {
		for _, b := range inOut.Get() {
			script.AddData(b)
		}
	}
	pkScript, err := script.Script()
	if err != nil {
		return nil, fmt.Errorf("error building token sell script; %w", err)
	}
	return pkScript, nil
}

func (t TokenSell) Type() memo.OutputType {
	return memo.OutputTypeTokenSell
}
