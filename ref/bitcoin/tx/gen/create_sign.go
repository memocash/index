package gen

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func (c *Create) Sign(msg *wire.MsgTx, keyRing wallet.KeyRing) error {
	if err := Sign(msg, c.getTxInputs(), keyRing); err != nil {
		return jerr.Get("error signing tx in create", err)
	}
	return nil
}
