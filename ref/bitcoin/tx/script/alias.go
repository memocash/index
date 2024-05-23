package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Alias struct {
	Hash  []byte
	Alias string
}

func (p Alias) Get() ([]byte, error) {
	if len(p.Hash) != memo.PkHashLength || len(p.Hash) != memo.ScriptHashLength {
		return nil, fmt.Errorf("pk hash incorrect length %d", len(p.Hash))
	}
	alias := []byte(p.Alias)
	if len(alias) > memo.OldMaxSendSize {
		return nil, fmt.Errorf("alias size too large")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSetAlias).
		AddData(p.Hash).
		AddData(alias).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building alias script; %w", err)
	}
	return pkScript, nil
}

func (p Alias) Type() memo.OutputType {
	return memo.OutputTypeSetAlias
}
