package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type FollowUser struct {
	UserPkHash []byte
	Unfollow   bool
}

func (t FollowUser) Get() ([]byte, error) {
	if len(t.UserPkHash) != memo.PkHashLength {
		return nil, jerr.New("incorrect pk hash length")
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
		return nil, jerr.Get("error building user follow script", err)
	}
	return pkScript, nil
}

func (t FollowUser) Type() memo.OutputType {
	if t.Unfollow {
		return memo.OutputTypeMemoUnfollow
	}
	return memo.OutputTypeMemoFollow
}
