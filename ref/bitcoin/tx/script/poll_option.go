package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type PollOption struct {
	PollTxHash []byte
	Option     string
}

func (v PollOption) Get() ([]byte, error) {
	option := []byte(v.Option)
	if len(v.PollTxHash) != memo.TxHashLength {
		return nil, fmt.Errorf("error invalid poll option tx hash length (%d)", len(v.PollTxHash))
	}
	if len(option) > memo.MaxPollOptionSize {
		return nil, fmt.Errorf("error option data too large")
	}
	if len(option) == 0 {
		return nil, fmt.Errorf("error option empty")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixPollOption).
		AddData(v.PollTxHash).
		AddData(option).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building poll option script; %w", err)
	}
	return pkScript, nil
}

func (v PollOption) Type() memo.OutputType {
	return memo.OutputTypeMemoPollOption
}
