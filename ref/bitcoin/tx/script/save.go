package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Save struct {
	Filename string
	Filetype string
	Contents []byte
}

func (s Save) Get() ([]byte, error) {
	if len(s.Contents) > memo.MaxFileSize {
		return nil, fmt.Errorf("file size too large")
	}
	if len(s.Contents) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixBitcom).
		AddData(s.Contents).
		AddData([]byte(s.Filetype)).
		AddData([]byte("")).
		AddData([]byte(s.Filename)).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building script; %w", err)
	}
	return pkScript, nil
}

func (s Save) Type() memo.OutputType {
	return memo.OutputTypeBitcom
}
