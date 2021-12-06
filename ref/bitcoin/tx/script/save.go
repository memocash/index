package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Save struct {
	Filename string
	Filetype string
	Contents []byte
}

func (s Save) Get() ([]byte, error) {
	if len(s.Contents) > memo.MaxFileSize {
		return nil, jerr.New("file size too large")
	}
	if len(s.Contents) == 0 {
		return nil, jerr.New("empty file")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixBitcom).
		AddData(s.Contents).
		AddData([]byte(s.Filetype)).
		AddData([]byte("")).
		AddData([]byte(s.Filename)).
		Script()
	if err != nil {
		return nil, jerr.Get("error building script", err)
	}
	return pkScript, nil
}

func (s Save) Type() memo.OutputType {
	return memo.OutputTypeBitcom
}
