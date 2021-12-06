package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type LikeRequest struct {
	Wallet     Wallet
	TxHash     []byte
	Tip        int64
	TipAddress wallet.Address
}

func Like(request LikeRequest) ([]*memo.Tx, error) {
	outputs := []*memo.Output{{
		Script: &script.Like{
			TxHash: request.TxHash,
		},
	}}
	if request.Tip != 0 {
		if request.Tip < memo.DustMinimumOutput {
			return nil, jerr.New("error tip not above dust limit")
		}
		if request.Tip > 1e8 {
			return nil, jerr.New("error trying to tip too much")
		}
		output := gen.GetAddressOutput(request.TipAddress, request.Tip)
		if output == nil {
			return nil, jerr.New(wallet.UnknownAddressTypeErrorMessage)
		}
		outputs = append(outputs, output)
	}
	txs, err := Simple(request.Wallet, outputs)
	if err != nil {
		return nil, jerr.Get("error building like tx", err)
	}
	return txs, nil
}
