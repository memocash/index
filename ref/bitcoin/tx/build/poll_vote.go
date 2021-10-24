package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

type PollVoteRequest struct {
	Wallet           Wallet
	PollOptionTxHash []byte
	Message          string
	Tip              int64
	TipAddress       wallet.Address
}

func PollVote(request PollVoteRequest) ([]*memo.Tx, error) {
	outputs := []*memo.Output{{
		Script: &script.PollVote{
			PollOptionTxHash: request.PollOptionTxHash,
			Message:          request.Message,
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
		return nil, jerr.Get("error building poll vote tx", err)
	}
	return txs, nil
}
