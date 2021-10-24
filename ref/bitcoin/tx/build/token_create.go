package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

type TokenCreateRequest struct {
	Wallet       Wallet
	Ticker       string
	Name         string
	Decimals     int
	DocUrl       string
	SlpType      byte
	Quantity     uint64
	BatonAddress wallet.Address
	TokenAddress wallet.Address
	NftUtxo      *memo.UTXO
}

func TokenCreate(request TokenCreateRequest) (*memo.Tx, error) {
	if request.SlpType == 0 {
		return nil, jerr.New("error empty token type")
	}
	if ! request.BatonAddress.IsSet() {
		request.BatonAddress = request.Wallet.GetSlpAddress()
	}
	if ! request.TokenAddress.IsSet() {
		request.TokenAddress = request.Wallet.GetSlpAddress()
	}
	outputs := []*memo.Output{{
		Script: &script.TokenCreate{
			Ticker:   request.Ticker,
			Name:     request.Name,
			SlpType:  request.SlpType,
			Decimals: request.Decimals,
			DocUrl:   request.DocUrl,
			Quantity: request.Quantity,
		},
	}}
	outputs = append(outputs, gen.GetAddressOutput(request.TokenAddress, memo.DustMinimumOutput))
	var inputsToUse []memo.UTXO
	if request.SlpType == memo.SlpNftChildTokenType {
		if request.NftUtxo == nil {
			return nil, jerr.New("nft child token set but nft group utxo not set")
		}
		inputsToUse = append(inputsToUse, *request.NftUtxo)
	} else {
		outputs = append(outputs, gen.GetAddressOutput(request.BatonAddress, memo.DustMinimumOutput))
	}
	tx, err := gen.Tx(gen.TxRequest{
		Getter:      request.Wallet.Getter,
		Outputs:     outputs,
		Change:      request.Wallet.GetChange(),
		InputsToUse: inputsToUse,
		KeyRing:     request.Wallet.KeyRing,
	})
	if err != nil {
		return nil, jerr.Get("error building token create tx", err)
	}
	return tx, nil
}
