package parse

import (
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jutil"
)

type SlpSend struct {
	TokenType  uint16
	TokenHash  []byte
	Quantities []uint64
}

func (c *SlpSend) Parse(pkScript []byte) error {
	pushData, err := txscript.PushedData(pkScript)
	if err != nil {
		return fmt.Errorf("error parsing pk script push data; %w", err)
	}
	const ExpectedPushDataCount = 5
	if len(pushData) < ExpectedPushDataCount {
		return fmt.Errorf("error invalid send, incorrect push data (%d), expected %d",
			len(pushData), ExpectedPushDataCount)
	}
	c.TokenType = uint16(jutil.GetUint64(pushData[1]))
	c.TokenHash = jutil.ByteReverse(pushData[3])
	c.Quantities = make([]uint64, len(pushData)-4)
	for i := 4; i < len(pushData); i++ {
		var index = i - 4
		c.Quantities[index] = jutil.GetUint64(pushData[i])
	}
	return nil
}

func NewSlpSend() *SlpSend {
	return &SlpSend{}
}
