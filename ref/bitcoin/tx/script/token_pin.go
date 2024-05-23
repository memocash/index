package script

import (
	"encoding/binary"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type TokenPin struct {
	PostTxHash  []byte
	TokenTxHash []byte
	TokenIndex  uint
}

func (t TokenPin) Get() ([]byte, error) {
	var indexBytes = make([]byte, 2)
	binary.BigEndian.PutUint16(indexBytes, uint16(t.TokenIndex))
	script := memo.GetBaseOpReturn().
		AddData(memo.PrefixTokenPin).
		AddData(jutil.ByteReverse(t.PostTxHash)).
		AddData(jutil.ByteReverse(t.TokenTxHash)).
		AddData(indexBytes)

	pkScript, err := script.Script()
	if err != nil {
		return nil, fmt.Errorf("error building token sell script; %w", err)
	}
	return pkScript, nil
}

func (t TokenPin) Type() memo.OutputType {
	return memo.OutputTypeTokenPin
}
