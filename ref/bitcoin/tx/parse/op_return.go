package parse

import (
	"github.com/jchavannes/btcd/wire"
	"time"
)

type OpReturn struct {
	Seen     time.Time
	Saved    bool
	TxHash   [32]byte
	Addr     [25]byte
	PushData [][]byte
	Outputs  []*wire.TxOut
}
