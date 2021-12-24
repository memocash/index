package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type TokenSendRequest struct {
	Wallet    Wallet
	TokenHash []byte
	Recipient wallet.Address
	Quantity  uint64
	TokenType byte
	NoSell    bool
}

func TokenSend(request TokenSendRequest) (*memo.Tx, error) {
	outputs := []*memo.Output{
		{Script: &script.TokenSend{
			TokenHash:  request.TokenHash,
			SlpType:    request.TokenType,
			Quantities: []uint64{request.Quantity},
		}},
		gen.GetAddressOutput(request.Recipient, memo.DustMinimumOutput),
	}
	tx, err := SimpleSingle(request.Wallet, outputs)
	if err != nil {
		return nil, jerr.Get("error building token send tx", err)
	}
	return tx, nil
}
