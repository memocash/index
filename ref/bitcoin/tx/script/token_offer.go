package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type TokenOffer struct {
	SellTxHash []byte
	InOuts     []InOut
}

func (t TokenOffer) Get() ([]byte, error) {
	script := memo.GetBaseOpReturn().
		AddData(memo.PrefixSellTokenOffer).
		AddData(t.SellTxHash)
	for _, inOut := range t.InOuts {
		for _, b := range inOut.Get() {
			script.AddData(b)
		}
	}
	pkScript, err := script.Script()
	if err != nil {
		return nil, jerr.Get("error building token offer script", err)
	}

	return pkScript, nil
}

func (t TokenOffer) Type() memo.OutputType {
	return memo.OutputTypeTokenOffer
}
