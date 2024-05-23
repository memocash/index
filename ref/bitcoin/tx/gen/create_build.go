package gen

import (
	"fmt"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

func (c *Create) Build() (*wire.MsgTx, error) {
	if c.Request.Getter != nil {
		c.Request.Getter.NewTx()
	}
	err := c.CheckOutputs()
	if err != nil {
		return nil, fmt.Errorf("error with outputs; %w", err)
	}
	for {
		err := c.setSlpInputs()
		if err != nil {
			return nil, fmt.Errorf("error setting slp inputs; %w", err)
		}
		err = c.setSlpChange()
		if err != nil {
			return nil, fmt.Errorf("error setting slp change; %w", err)
		}
		err = c.setInputs()
		if err != nil {
			return nil, fmt.Errorf("error setting inputs; %w", err)
		}
		err = c.setChange()
		if err != nil {
			return nil, fmt.Errorf("error setting change; %w", err)
		}
		if c.isComplete() {
			break
		}
		if c.Request.Getter == nil {
			return nil, fmt.Errorf("request getter not set; %w", c.getNotEnoughValueError())
		}
		moreUTXOs, err := c.getMoreUTXOs()
		if err != nil {
			return nil, fmt.Errorf("error getting more utxos; %w", err)
		}
		if len(moreUTXOs) == 0 {
			if !c.isEnoughSlpValue() {
				return nil, fmt.Errorf("error no more utxos and not enough slp value; %w", c.getNotEnoughTokenValueError())
			}
			enoughValue, err := c.isEnoughInputValue()
			if err != nil {
				return nil, fmt.Errorf("error determining if enough value; %w", err)
			}
			if !enoughValue {
				return nil, fmt.Errorf("error no more utxos and not enough value; %w", c.getNotEnoughValueError())
			}
			return nil, fmt.Errorf("error no more utxos, tx incomplete, unknown error")
		}
		c.PotentialInputs = append(c.PotentialInputs, moreUTXOs...)
	}
	wireTx, err := c.getWireTx()
	if err != nil {
		return nil, fmt.Errorf("error getting wire tx; %w", err)
	}
	return wireTx, nil
}

func (c Create) CheckOutputs() error {
	for _, output := range c.Outputs {
		switch output.Script.(type) {
		case *script.P2pkh, *script.P2sh:
		default:
			continue
		}
		if output.Amount < memo.DustMinimumOutput {
			return fmt.Errorf("error output below dust limit (type: %s, amount: %d); %w",
				output.GetType(), output.Amount, BelowDustLimitError)
		}
	}
	return nil
}
