package gen

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

type MultiRequest struct {
	Outputs      []*memo.Output
	Getter       InputGetter
	FaucetGetter InputGetter
	FaucetSaver  FaucetSaver
	InputsToUse  []memo.UTXO
	KeyRing      wallet.KeyRing
	Change       wallet.Change
}

func Multi(request MultiRequest) ([]*memo.Tx, error) {
	memoTx, err := Tx(TxRequest{
		Getter:      request.Getter,
		Outputs:     request.Outputs,
		Change:      request.Change,
		InputsToUse: request.InputsToUse,
		KeyRing:     request.KeyRing,
	})
	if err == nil {
		return []*memo.Tx{memoTx}, nil
	} else if ! IsNotEnoughValueError(err) || jutil.IsNil(request.FaucetSaver) ||
		! request.FaucetSaver.IsFreeTx(request.Outputs) || script.IsBigMemo(request.Outputs) ||
		jutil.IsNil(request.FaucetGetter) {
		return nil, jerr.Get("error generating tx", err)
	}
	faucetKey := request.FaucetSaver.GetKey()
	faucetTx, utxo, err := FaucetTx(request.Change.Main.GetPkHash(), request.FaucetGetter, faucetKey)
	if err != nil {
		return nil, jerr.Get("error getting faucet tx", err)
	}
	faucetAddress := faucetKey.GetPublicKey().GetAddress()
	amount := utxo.Input.Value - memo.FreeTxFee(request.Outputs)
	memoTx, err = Tx(TxRequest{
		InputsToUse: append(request.InputsToUse, utxo),
		Outputs:     append(request.Outputs, GetAddressOutput(faucetAddress, amount)),
		Change:      request.Change,
		KeyRing:     request.KeyRing,
	})
	if err != nil {
		return nil, jerr.Get("error generating tx", err)
	}
	if request.FaucetSaver != nil {
		err = request.FaucetSaver.Save(
			request.Change.Main.GetPkHash(),
			faucetKey.GetPkHash(),
			faucetTx.GetHash(),
			memoTx.GetHash(),
		)
		if err != nil {
			return nil, jerr.Get("error saving faucet transaction", err)
		}
	}
	return []*memo.Tx{
		faucetTx,
		memoTx,
	}, nil
}
