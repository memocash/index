package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type Signature struct {
	Sig    []byte
	PkData []byte
}

type TokenSignature struct {
	OfferTxHash []byte
	Signatures  []Signature
}

func (t TokenSignature) Get() ([]byte, error) {
	script := memo.GetBaseOpReturn().
		AddData(memo.PrefixSellTokenSignature).
		AddData(t.OfferTxHash)
	for _, signature := range t.Signatures {
		script = script.
			AddData(signature.Sig).
			AddData(signature.PkData)
	}
	pkScript, err := script.Script()
	if err != nil {
		return nil, jerr.Get("error building token signature script", err)
	}
	return pkScript, nil
}

func (t TokenSignature) Type() memo.OutputType {
	return memo.OutputTypeTokenSignature
}
