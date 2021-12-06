package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type TokenMintRequest struct {
	Wallet       Wallet
	Baton        memo.UTXO
	BatonAddress wallet.Address
	TokenAddress wallet.Address
	TokenHash    []byte
	TokenType    byte
	Quantity     uint64
}

func TokenMint(request TokenMintRequest) (*memo.Tx, error) {
	tx, err := gen.Tx(gen.TxRequest{
		Getter: request.Wallet.Getter,
		Outputs: []*memo.Output{{
			Script: &script.TokenMint{
				TokenHash: request.TokenHash,
				TokenType: request.TokenType,
				Quantity:  request.Quantity,
			}},
			gen.GetAddressOutput(request.TokenAddress, memo.DustMinimumOutput),
			gen.GetAddressOutput(request.BatonAddress, memo.DustMinimumOutput),
		},
		Change:      request.Wallet.GetChange(),
		InputsToUse: []memo.UTXO{request.Baton},
		KeyRing:     request.Wallet.KeyRing,
	})
	if err != nil {
		return nil, jerr.Get("error building token mint tx", err)
	}
	return tx, nil
}
