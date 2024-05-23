package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type TokenSellAcceptRequest struct {
	Wallet     Wallet
	TokenInput memo.UTXO
	PayAddress wallet.Address
	PayAmount  int64
	TokenHash  []byte
	TokenType  byte
	TokenAmt   uint64
	Signature  []byte
	PkData     []byte
	FeeAddress wallet.Address
	Fee        int64
}

func TokenSellAccept(request TokenSellAcceptRequest) (*memo.Tx, error) {
	var tokenSendOutput = &memo.Output{
		Script: &script.TokenSend{
			TokenHash:  request.TokenHash,
			SlpType:    request.TokenType,
			Quantities: []uint64{0, request.TokenAmt},
		},
	}
	var outputs = []*memo.Output{
		tokenSendOutput,
		gen.GetAddressOutput(request.PayAddress, request.PayAmount),
		gen.GetAddressOutput(request.Wallet.GetSlpAddress(), memo.DustMinimumOutput),
	}
	if request.Fee > 0 && request.FeeAddress.IsSet() {
		outputs = append(outputs, gen.GetAddressOutput(request.FeeAddress, request.Fee))
	}
	sellAcceptTx, err := gen.TxUnsigned(gen.TxRequest{
		Getter:      request.Wallet.Getter,
		InputsToUse: []memo.UTXO{request.TokenInput},
		Outputs:     outputs,
		Change:      request.Wallet.GetChange(),
		KeyRing:     request.Wallet.KeyRing,
	})
	if err != nil {
		return nil, fmt.Errorf("error building unsigned sell accept tx; %w", err)
	}
	if len(sellAcceptTx.Inputs) < 2 {
		return nil, fmt.Errorf("error sell accept inputs less than 2 (%d)", len(sellAcceptTx.Inputs))
	} else {
		sellAcceptTx.Inputs[0], sellAcceptTx.Inputs[1] = sellAcceptTx.Inputs[1], sellAcceptTx.Inputs[0]
		sellAcceptTx.MsgTx.TxIn[0], sellAcceptTx.MsgTx.TxIn[1] = sellAcceptTx.MsgTx.TxIn[1], sellAcceptTx.MsgTx.TxIn[0]
	}
	err = gen.AttachSignatureToInput(sellAcceptTx.MsgTx.TxIn[1], request.Signature, request.PkData)
	if err != nil {
		return nil, fmt.Errorf("error attaching sell signature to transaction input; %w", err)
	}
	err = gen.Sign(sellAcceptTx.MsgTx, gen.GetNonPointerTxInputs(sellAcceptTx.Inputs), request.Wallet.KeyRing)
	if err != nil {
		return nil, fmt.Errorf("error signing rest of token sell accept tx; %w", err)
	}
	return sellAcceptTx, nil
}
