package parse

import (
	"github.com/jchavannes/btcd/wire"
)

type OpReturn struct {
	Height   int64
	TxHash   [32]byte
	Addr     [25]byte
	PushData [][]byte
	Outputs  []*wire.TxOut
}
