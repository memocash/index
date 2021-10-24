package parse

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
)

type SlpCreate struct {
	TokenType  byte
	Ticker     string
	Name       string
	DocUrl     string
	DocHash    []byte
	Decimals   uint8
	BatonIndex uint32
	Quantity   uint64
}

func (c *SlpCreate) Parse(pkScript []byte) error {
	pushData, err := txscript.PushedData(pkScript)
	if err != nil {
		return jerr.Get("error parsing pk script push data", err)
	}
	const ExpectedPushDataCount = 10
	if len(pushData) < ExpectedPushDataCount {
		return jerr.Newf("invalid genesis, incorrect push data (%d), expected %d", len(pushData), ExpectedPushDataCount)
	}
	c.TokenType = byte(jutil.GetUint64(pushData[1]))
	c.Ticker = jutil.GetUtf8String(pushData[3])
	c.Name = jutil.GetUtf8String(pushData[4])
	c.DocUrl = jutil.GetUtf8String(pushData[5])
	c.DocHash = pushData[6]
	c.Decimals = uint8(jutil.GetUint64(pushData[7]))
	c.BatonIndex = uint32(jutil.GetUint64(pushData[8]))
	c.Quantity = jutil.GetUint64(pushData[9])
	return nil
}

func NewSlpCreate() *SlpCreate {
	return &SlpCreate{}
}
