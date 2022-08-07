package parse

import (
	"github.com/jchavannes/btcd/wire"
)

type OpReturn struct {
	Height   int64
	TxHash   []byte
	LockHash []byte
	PushData [][]byte
	Outputs  []*wire.TxOut
}
