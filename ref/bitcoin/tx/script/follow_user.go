package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type FollowUser struct {
	UserPkHash []byte
	Unfollow   bool
}

func (t FollowUser) Get() ([]byte, error) {
	if len(t.UserPkHash) != memo.PkHashLength {
		return nil, fmt.Errorf("incorrect pk hash length")
	}
	var prefix = memo.PrefixFollow
	if t.Unfollow {
		prefix = memo.PrefixUnfollow
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(prefix).
		AddData(t.UserPkHash).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building user follow script; %w", err)
	}
	return pkScript, nil
}

func (t FollowUser) Type() memo.OutputType {
	if t.Unfollow {
		return memo.OutputTypeMemoUnfollow
	}
	return memo.OutputTypeMemoFollow
}
