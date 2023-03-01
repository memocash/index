package save

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

func SlpGenesis(info parse.OpReturn) error {
	const ExpectedPushDataCount = 10
	if len(info.PushData) < ExpectedPushDataCount {
		return jerr.Newf("invalid genesis, incorrect push data (%d), expected %d", len(info.PushData), ExpectedPushDataCount)
	}
	docHash, _ := chainhash.NewHash(info.PushData[6])
	var genesis = &slp.Genesis{
		TxHash:     info.TxHash,
		TokenType:  uint8(jutil.GetUint16(info.PushData[1])),
		Ticker:     jutil.GetUtf8String(info.PushData[3]),
		Name:       jutil.GetUtf8String(info.PushData[4]),
		DocUrl:     jutil.GetUtf8String(info.PushData[5]),
		DocHash:    *docHash,
		Decimals:   uint8(jutil.GetUint64(info.PushData[7])),
		BatonIndex: uint32(jutil.GetUint64(info.PushData[8])),
		Quantity:   jutil.GetUint64(info.PushData[9]),
	}
	if err := db.Save([]db.Object{genesis}); err != nil {
		return jerr.Get("error saving slp genesis", err)
	}
	if err := SlpOutput(info, genesis.TxHash, memo.SlpMintTokenIndex, genesis.Quantity); err != nil {
		return jerr.Get("error saving slp output", err)
	}
	return nil
}

func SlpMint(info parse.OpReturn) error {
	return nil
}

func SlpSend(info parse.OpReturn) error {
	return nil
}

func SlpCommit(info parse.OpReturn) error {
	return nil
}

func SlpOutput(info parse.OpReturn, tokenHash [32]byte, index uint32, quantity uint64) error {
	if quantity == 0 {
		return nil
	}
	if len(info.Outputs) <= int(index) {
		return jerr.Newf("slp tx out index out of range (len: %d, index: %d)", len(info.Outputs), index)
	}
	if err := db.Save([]db.Object{&slp.Output{
		TxHash:    info.TxHash,
		Index:     index,
		TokenHash: tokenHash,
		Quantity:  quantity,
	}}); err != nil {
		return jerr.Get("error saving slp output", err)
	}
	return nil
}
