package gen

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type InputGetterOld interface {
	SetPkHashesToUse([][]byte)
	GetUTXOsOld(*memo.UTXORequest) ([]memo.UTXO, error)
}

type InputGetterWrapper struct {
	Old      InputGetterOld
	PkHashes [][]byte
	UTXOs    []memo.UTXO
	Used     []memo.UTXO
	pksToUse [][]byte
	reset    bool
}

func GetWrapper(getter InputGetterOld, pkHash []byte) InputGetter {
	return &InputGetterWrapper{
		Old:      getter,
		PkHashes: [][]byte{pkHash},
	}
}

func GetWrapperMultiKey(getter InputGetterOld, pkHashes [][]byte) InputGetter {
	return &InputGetterWrapper{
		Old:      getter,
		PkHashes: pkHashes,
	}
}

func (w InputGetterWrapper) GetPkHashes() [][]byte {
	var pkHashes = w.PkHashes
	if len(w.pksToUse) > 0 {
	loop:
		for i := 0; i < len(pkHashes); i++ {
			for _, pkToUse := range w.pksToUse {
				if bytes.Equal(pkToUse, pkHashes[i]) {
					continue loop
				}
			}
			pkHashes = append(pkHashes[:i], pkHashes[i+1:]...)
			i--
		}
	}
	return pkHashes
}

func (w *InputGetterWrapper) GetUTXOs(request *memo.UTXORequest) ([]memo.UTXO, error) {
	if request == nil {
		request = new(memo.UTXORequest)
	}
	pkHashes := w.GetPkHashes()
	var usableUtxos []memo.UTXO
utxoLoop:
	for _, utxo := range w.UTXOs {
		for _, pkHash := range pkHashes {
			if bytes.Equal(pkHash, utxo.Input.PkHash) {
				usableUtxos = append(usableUtxos, utxo)
				continue utxoLoop
			}
		}
	}
	if w.reset {
		w.reset = false
		if len(usableUtxos) > 0 {
			return usableUtxos, nil
		}
	}
	utxos, err := w.Old.GetUTXOsOld(request)
	if err != nil {
		return nil, jerr.Get("error getting UTXOs using wrapper", err)
	}
	for i := 0; i < len(utxos); i++ {
		for _, used := range w.Used {
			if bytes.Equal(used.Input.PrevOutHash, utxos[i].Input.PrevOutHash) &&
				used.Input.PrevOutIndex == utxos[i].Input.PrevOutIndex {
				utxos = append(utxos[:i], utxos[i+1:]...)
				i--
				break
			}
		}
	}
	w.UTXOs = append(w.UTXOs, utxos...)
	return utxos, nil
}

func (w *InputGetterWrapper) MarkUTXOsUsed(utxos []memo.UTXO) {
	w.Used = append(w.Used, utxos...)
loop:
	for i := 0; i < len(w.UTXOs); i++ {
		for _, utxo := range utxos {
			if utxo.IsEqual(w.UTXOs[i]) {
				w.UTXOs = append(w.UTXOs[:i], w.UTXOs[i+1:]...)
				i--
				continue loop
			}
		}
	}
}

func (w *InputGetterWrapper) SetPkHashesToUse(pkHashes [][]byte) {
	w.pksToUse = pkHashes
	w.Old.SetPkHashesToUse(pkHashes)
}

func (w *InputGetterWrapper) AddChangeUTXO(utxo memo.UTXO) {
	for _, pkHash := range w.PkHashes {
		if bytes.Equal(pkHash, utxo.Input.PkHash) {
			utxo.AncestorsNC++
			w.UTXOs = append([]memo.UTXO{utxo}, w.UTXOs...)
			return
		}
	}
}

func (w *InputGetterWrapper) NewTx() {
	w.reset = true
}

type InputGetter interface {
	SetPkHashesToUse([][]byte)
	GetUTXOs(*memo.UTXORequest) ([]memo.UTXO, error)
	MarkUTXOsUsed([]memo.UTXO)
	AddChangeUTXO(memo.UTXO)
	NewTx()
}

type FaucetSaver interface {
	Save(userPkHash []byte, faucetPkHash []byte, fundTxHash []byte, memoTxHash []byte) error
	IsFreeTx([]*memo.Output) bool
	GetKey() wallet.PrivateKey
}

type TxRequest struct {
	Getter      InputGetter
	InputsToUse []memo.UTXO
	Outputs     []*memo.Output
	Change      wallet.Change
	KeyRing     wallet.KeyRing
}

func (r TxRequest) GetTokenSendOutput() *script.TokenSend {
	for _, spendOutput := range r.Outputs {
		spendOutputScript, ok := spendOutput.Script.(*script.TokenSend)
		if ok {
			return spendOutputScript
		}
	}
	return nil
}

func (r TxRequest) GetTokenHash() []byte {
	for _, spendOutput := range r.Outputs {
		switch v := spendOutput.Script.(type) {
		case *script.TokenSend:
			return v.TokenHash
		case *script.TokenMint:
			return v.TokenHash
		}
	}
	return nil
}
