package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
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
		return nil, fmt.Errorf("error building token sell tx; %w", err)
	}
	return tx, nil
}
