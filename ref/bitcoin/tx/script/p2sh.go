package script

import (
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type P2sh struct {
	ScriptHash []byte
}

func (p P2sh) Get() ([]byte, error) {
	if len(p.ScriptHash) != memo.ScriptHashLength {
		return nil, fmt.Errorf("invalid script hash length: %d (expected %d)", len(p.ScriptHash), memo.ScriptHashLength)
	}
	pkScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_HASH160).
		AddData(p.ScriptHash).
		AddOp(txscript.OP_EQUAL).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building p2sh script; %w", err)
	}
	return pkScript, nil
}

func (p P2sh) Type() memo.OutputType {
	return memo.OutputTypeP2SH
}
