package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
)

type TokenSellSignatureRequest struct {
	Wallet      Wallet
	OfferTxHash []byte
	Signatures  []script.Signature
}

func TokenSellSignature(request TokenSellSignatureRequest) (*memo.Tx, error) {
	tx, err := SimpleSingle(request.Wallet, []*memo.Output{{
		Script: &script.TokenSignature{
			OfferTxHash: request.OfferTxHash,
			Signatures:  request.Signatures,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building token sell tx", err)
	}
	return tx, nil
}
