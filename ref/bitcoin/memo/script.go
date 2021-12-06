package memo

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Script interface {
	Get() ([]byte, error)
	Type() OutputType
}

var _chain string

func SetChain(chain string) {
	_chain = chain
}

func GetBaseOpReturn() *txscript.ScriptBuilder {
	builder := txscript.NewScriptBuilder()
	if _chain == wallet.ChainNameSV {
		builder = builder.AddData([]byte{})
	}
	return builder.AddOp(txscript.OP_RETURN)
}
