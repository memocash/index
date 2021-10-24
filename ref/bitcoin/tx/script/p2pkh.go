package script

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type P2pkh struct {
	PkHash []byte
}

func (p P2pkh) Get() ([]byte, error) {
	if len(p.PkHash) != memo.PkHashLength {
		return nil, jerr.Newf("invalid pk hash length: %d (expected %d)", len(p.PkHash), memo.PkHashLength)
	}
	pkScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(p.PkHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
	if err != nil {
		return nil, jerr.Get("error building p2pkh script", err)
	}
	return pkScript, nil
}

func (p P2pkh) Type() memo.OutputType {
	return memo.OutputTypeP2PKH
}
