package gen

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

func (c *Create) Sign(msg *wire.MsgTx, keyRing wallet.KeyRing) error {
	err := Sign(msg, c.getTxInputs(), keyRing)
	if err != nil {
		return jerr.Get("error signing tx", err)
	}
	return nil
}
