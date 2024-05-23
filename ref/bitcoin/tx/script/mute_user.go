package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type MuteUser struct {
	MutePkHash []byte
	Unmute     bool
}

func (m MuteUser) Get() ([]byte, error) {
	if len(m.MutePkHash) != memo.PkHashLength {
		return nil, fmt.Errorf("invalid address length")
	}
	var prefix []byte
	if m.Unmute {
		prefix = memo.PrefixUnmute
	} else {
		prefix = memo.PrefixMute
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(prefix).
		AddData(m.MutePkHash).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building mute user script; %w", err)
	}
	return pkScript, nil
}

func (m MuteUser) Type() memo.OutputType {
	if m.Unmute {
		return memo.OutputTypeMemoUnMute
	}
	return memo.OutputTypeMemoMute
}
