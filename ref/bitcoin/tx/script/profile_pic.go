package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type ProfilePic struct {
	Url string
}

func (p ProfilePic) Get() ([]byte, error) {
	url := []byte(p.Url)
	if len(url) > memo.OldMaxPostSize {
		return nil, jerr.New("url size too large")
	}
	if len(url) == 0 {
		return nil, jerr.New("empty url")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSetProfilePic).
		AddData(url).
		Script()
	if err != nil {
		return nil, jerr.Get("error building script", err)
	}
	return pkScript, nil
}

func (p ProfilePic) Type() memo.OutputType {
	return memo.OutputTypeMemoSetProfilePic
}
