package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type ProfilePicRequest struct {
	Wallet Wallet
	Url    string
}

func ProfilePic(request ProfilePicRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.ProfilePic{
			Url: request.Url,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building set profile pic tx", err)
	}
	return txs, nil
}
