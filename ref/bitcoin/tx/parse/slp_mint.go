package parse

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
)

type SlpMint struct {
	TokenType  uint16
	TokenHash  []byte
	BatonIndex uint32
	Quantity   uint64
}

func (c *SlpMint) Parse(pkScript []byte) error {
	pushData, err := txscript.PushedData(pkScript)
	if err != nil {
		return jerr.Get("error parsing pk script push data", err)
	}
	const ExpectedPushDataCount = 6
	if len(pushData) < ExpectedPushDataCount {
		return jerr.Newf("error invalid mint, incorrect push data (%d), expected %d",
			len(pushData), ExpectedPushDataCount)
	}
	c.TokenType = uint16(jutil.GetUint64(pushData[1]))
	c.TokenHash = jutil.ByteReverse(pushData[3])
	c.BatonIndex = uint32(jutil.GetUint64(pushData[4]))
	c.Quantity = jutil.GetUint64(pushData[5])
	return nil
}

func NewSlpMint() *SlpMint {
	return &SlpMint{}
}
