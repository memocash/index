package gen

import (
	"fmt"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func (c *Create) Sign(msg *wire.MsgTx, keyRing wallet.KeyRing) error {
	if err := Sign(msg, c.getTxInputs(), keyRing); err != nil {
		return fmt.Errorf("error signing tx in create; %w", err)
	}
	return nil
}
