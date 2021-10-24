package gen

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

func (c *Create) Build() (*wire.MsgTx, error) {
	if c.Request.Getter != nil {
		c.Request.Getter.NewTx()
	}
	err := c.CheckOutputs()
	if err != nil {
		return nil, jerr.Get("error with outputs", err)
	}
	for {
		err := c.setSlpInputs()
		if err != nil {
			return nil, jerr.Get("error setting slp inputs", err)
		}
		err = c.setSlpChange()
		if err != nil {
			return nil, jerr.Get("error setting slp change", err)
		}
		err = c.setInputs()
		if err != nil {
			return nil, jerr.Get("error setting inputs", err)
		}
		err = c.setChange()
		if err != nil {
			return nil, jerr.Get("error setting change", err)
		}
		if c.isComplete() {
			break
		}
		if c.Request.Getter == nil {
			return nil, jerr.Get("request getter not set", c.getNotEnoughValueError())
		}
		moreUTXOs, err := c.getMoreUTXOs()
		if err != nil {
			return nil, jerr.Get("error getting more utxos", err)
		}
		if len(moreUTXOs) == 0 {
			if ! c.isEnoughSlpValue() {
				return nil, jerr.Get("error no more utxos and not enough slp value", c.getNotEnoughTokenValueError())
			}
			enoughValue, err := c.isEnoughInputValue()
			if err != nil {
				return nil, jerr.Get("error determining if enough value", err)
			}
			if ! enoughValue {
				return nil, jerr.Get("error no more utxos and not enough value", c.getNotEnoughValueError())
			}
			return nil, jerr.New("error no more utxos, tx incomplete, unknown error")
		}
		c.PotentialInputs = append(c.PotentialInputs, moreUTXOs...)
	}
	wireTx, err := c.getWireTx()
	if err != nil {
		return nil, jerr.Get("error getting wire tx", err)
	}
	return wireTx, nil
}

func (c Create) CheckOutputs() error {
	for _, output := range c.Outputs {
		if output.GetType() != memo.OutputTypeP2PKH && output.GetType() != memo.OutputTypeP2SH {
			continue
		}
		if output.Amount < memo.DustMinimumOutput {
			return jerr.Getf(BelowDustLimitError, "error output below dust limit (type: %s, amount: %d)",
				output.GetType(), output.Amount)
		}
	}
	return nil
}
