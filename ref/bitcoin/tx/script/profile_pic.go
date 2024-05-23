package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type ProfilePic struct {
	Url string
}

func (p ProfilePic) Get() ([]byte, error) {
	url := []byte(p.Url)
	if len(url) > memo.OldMaxPostSize {
		return nil, fmt.Errorf("url size too large")
	}
	if len(url) == 0 {
		return nil, fmt.Errorf("empty url")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixSetProfilePic).
		AddData(url).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building script; %w", err)
	}
	return pkScript, nil
}

func (p ProfilePic) Type() memo.OutputType {
	return memo.OutputTypeMemoSetProfilePic
}
