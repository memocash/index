package gen

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

func FaucetTx(pkHash []byte, faucetGetter InputGetter, faucetKey wallet.PrivateKey) (*memo.Tx, memo.UTXO, error) {
	address, err := wallet.GetAddressFromPkHashNew(pkHash)
	if err != nil {
		return nil, memo.UTXO{}, jerr.Get("error getting address from pk hash", err)
	}
	utxos, err := faucetGetter.GetUTXOs(nil)
	if err != nil {
		return nil, memo.UTXO{}, jerr.Get("error getting faucet utxos", err)
	}
	if len(utxos) == 0 {
		return nil, memo.UTXO{}, jerr.Get("insufficient funds in faucet", NotEnoughValueError)
	}
	var amount int64
	for _, utxo := range utxos {
		amount += utxo.Input.Value
	}
	var fee = memo.FeeP2pkh1In1OutTx + int64(len(utxos)-1)*memo.InputFeeP2PKH
	if amount > memo.MaxFundAmount {
		amount = jutil.MinInt64(memo.MaxFundAmount, amount/2)
		fee += memo.OutputFeeP2PKH
	}
	amount -= fee
	faucetTx, err := Tx(TxRequest{
		InputsToUse: utxos,
		Outputs: []*memo.Output{
			GetAddressOutput(address, amount),
		},
		Change:  wallet.GetChange(faucetKey.GetAddress()),
		KeyRing: wallet.GetSingleKeyRing(faucetKey),
	})
	if err != nil {
		return nil, memo.UTXO{}, jerr.Get("error generating faucet tx", err)
	}
	var utxo memo.UTXO
	for _, output := range script.GetOutputs(faucetTx) {
		if bytes.Equal(output.PkHash, pkHash) {
			utxo = memo.UTXO{
				Input:       *output,
				AncestorsNC: faucetTx.AncestorsNC,
			}
		}
	}
	return faucetTx, utxo, nil
}
