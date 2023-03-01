package save

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/slp"
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
		Addr:       info.Addr,
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
	/*output, err := saveSlpOutput(info, memo.SlpTxTypeGenesis, genesis.TxHash, memo.SlpMintTokenIndex, genesis.PkHash, genesis.Quantity)
	if err != nil {
		return jerr.Get("error saving slp output", err)
	}
	if genesis.TokenType == memo.SlpNftChildTokenType {
		for _, txIn := range info.Txn.TxIn {
			_, err = saveSlpInput(txIn)
			if err != nil {
				return jerr.Get("error saving slp input", err)
			}
		}
	}*/
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
