package save

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

func SlpGenesis(info parse.OpReturn) error {
	const ExpectedPushDataCount = 10
	if len(info.PushData) < ExpectedPushDataCount {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid slp genesis, incorrect push data (%d), expected %d", len(info.PushData), ExpectedPushDataCount),
		}); err != nil {
			return jerr.Get("error saving process error for slp genesis incorrect push data", err)
		}
		return nil
	}
	docHash, err := chainhash.NewHash(info.PushData[6])
	if err != nil {
		docHash = &chainhash.Hash{}
	}
	var genesis = &slp.Genesis{
		TxHash:     info.TxHash,
		TokenType:  uint8(jutil.GetUint64(info.PushData[1])),
		Ticker:     jutil.GetUtf8String(info.PushData[3]),
		Name:       jutil.GetUtf8String(info.PushData[4]),
		DocUrl:     jutil.GetUtf8String(info.PushData[5]),
		DocHash:    *docHash,
		Decimals:   uint8(jutil.GetUint64(info.PushData[7])),
		BatonIndex: uint32(jutil.GetUint64(info.PushData[8])),
	}
	if err := db.Save([]db.Object{genesis}); err != nil {
		return jerr.Get("error saving slp genesis op return to db", err)
	}
	if err := SlpOutput(info, genesis.TxHash, memo.SlpMintTokenIndex, jutil.GetUint64(info.PushData[9])); err != nil {
		return jerr.Get("error saving slp output for genesis", err)
	}
	if err := SlpBaton(info, genesis.TxHash, genesis.BatonIndex); err != nil {
		return jerr.Get("error saving slp baton for genesis", err)
	}
	return nil
}

func SlpMint(info parse.OpReturn) error {
	const ExpectedPushDataCount = 6
	if len(info.PushData) < ExpectedPushDataCount {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid slp mint, incorrect push data (%d), expected %d", len(info.PushData), ExpectedPushDataCount),
		}); err != nil {
			return jerr.Get("error saving process error for slp mint incorrect push data", err)
		}
		return nil
	}
	tokenHash, err := chainhash.NewHash(jutil.ByteReverse(info.PushData[3]))
	if err != nil {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid token hash for slp mint (%x)", info.PushData[3]),
		}); err != nil {
			return jerr.Get("error saving process error for slp mint invalid token hash", err)
		}
	}
	var mint = &slp.Mint{
		TxHash:     info.TxHash,
		TokenHash:  *tokenHash,
		BatonIndex: uint32(jutil.GetUint64(info.PushData[4])),
		Quantity:   jutil.GetUint64(info.PushData[5]),
	}
	if err := db.Save([]db.Object{mint}); err != nil {
		return jerr.Get("error saving mint op return to db", err)
	}
	if err := SlpOutput(info, mint.TokenHash, memo.SlpMintTokenIndex, mint.Quantity); err != nil {
		return jerr.Get("error saving slp output for mint", err)
	}
	if err := SlpBaton(info, mint.TokenHash, mint.BatonIndex); err != nil {
		return jerr.Get("error saving slp baton for mint", err)
	}
	return nil
}

func SlpSend(info parse.OpReturn) error {
	const ExpectedPushDataCount = 5
	if len(info.PushData) < ExpectedPushDataCount {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error: fmt.Sprintf("invalid slp send, incorrect push data (%d), expected %d",
				len(info.PushData), ExpectedPushDataCount)}); err != nil {
			return jerr.Get("error saving process error for slp send incorrect push data", err)
		}
		return nil
	}
	tokenHash, err := chainhash.NewHash(jutil.ByteReverse(info.PushData[3]))
	if err != nil {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid token hash for slp send (%x)", info.PushData[3]),
		}); err != nil {
			return jerr.Get("error saving process error for slp send invalid token hash", err)
		}
		return nil
	}
	var send = &slp.Send{
		TxHash:    info.TxHash,
		TokenHash: *tokenHash,
	}
	if err := db.Save([]db.Object{send}); err != nil {
		return jerr.Get("error saving send op return to db", err)
	}
	for i := 4; i < len(info.PushData); i++ {
		var index = uint32(i - 3)
		var quantity = jutil.GetUint64(info.PushData[i])
		if quantity == 0 {
			continue
		}
		if err := SlpOutput(info, send.TokenHash, index, quantity); err != nil {
			return jerr.Get("error saving slp output for send", err)
		}
	}
	return nil
}

func SlpCommit(parse.OpReturn) error {
	// Ignore commits for now
	return nil
}

func SlpOutput(info parse.OpReturn, tokenHash [32]byte, index uint32, quantity uint64) error {
	if quantity == 0 {
		return nil
	}
	if len(info.Outputs) <= int(index) {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error: fmt.Sprintf("invalid slp output, index out of range (len: %d, index: %d)",
				len(info.Outputs), index)}); err != nil {
			return jerr.Get("error saving process error for slp output index out of range", err)
		}
		return nil
	}
	if err := db.Save([]db.Object{&slp.Output{
		TxHash:    info.TxHash,
		Index:     index,
		TokenHash: tokenHash,
		Quantity:  quantity,
	}}); err != nil {
		return jerr.Get("error saving slp output op return to db", err)
	}
	return nil
}

func SlpBaton(info parse.OpReturn, tokenHash [32]byte, index uint32) error {
	if len(info.Outputs) <= int(index) {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error: fmt.Sprintf("invalid slp baton, index out of range (len: %d, index: %d)",
				len(info.Outputs), index)}); err != nil {
			return jerr.Get("error saving process error for slp baton index out of range", err)
		}
		return nil
	}
	if err := db.Save([]db.Object{&slp.Baton{
		TxHash:    info.TxHash,
		Index:     index,
		TokenHash: tokenHash,
	}}); err != nil {
		return jerr.Get("error saving slp baton op return to db", err)
	}
	return nil
}
