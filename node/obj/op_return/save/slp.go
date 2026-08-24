package save

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

// TranscribeSlp runs the lenient SLP transcription for one SLP op_return
// output. Lives in the save package (rather than op_return) so the validity
// cascade can transcribe reconstructed spenders without an import cycle;
// every path that writes a verdict must transcribe first, since validation
// treats a decided parent's missing output rows as definitive.
func TranscribeSlp(info parse.OpReturn) error {
	objects, err := TranscribeSlpObjects(info)
	if err != nil {
		return fmt.Errorf("error getting slp transcription objects; %w", err)
	}
	if len(objects) == 0 {
		return nil
	}
	if err := db.Save(objects); err != nil {
		return fmt.Errorf("error saving slp transcription objects; %w", err)
	}
	return nil
}

// TranscribeSlpObjects builds the db objects for one SLP op_return without
// saving them, so bulk callers can batch a single db.Save across many txs.
// Process errors for malformed messages are still logged immediately.
func TranscribeSlpObjects(info parse.OpReturn) ([]db.Object, error) {
	if len(info.PushData) < 5 {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid slp, incorrect push data (%d) op return handler", len(info.PushData)),
		}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp incorrect push data; %w", err)
		}
		return nil, nil
	}
	switch memo.SlpType(info.PushData[2]) {
	case memo.SlpTxTypeGenesis:
		objects, err := slpGenesisObjects(info)
		if err != nil {
			return nil, fmt.Errorf("error building slp genesis op return handler; %w", err)
		}
		return objects, nil
	case memo.SlpTxTypeMint:
		objects, err := slpMintObjects(info)
		if err != nil {
			return nil, fmt.Errorf("error building slp mint op return handler; %w", err)
		}
		return objects, nil
	case memo.SlpTxTypeSend:
		objects, err := slpSendObjects(info)
		if err != nil {
			return nil, fmt.Errorf("error building slp send op return handler; %w", err)
		}
		return objects, nil
	case memo.SlpTxTypeCommit:
		// Ignore commits for now
		return nil, nil
	default:
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("unknown slp tx type op return handler: %s", info.PushData[2]),
		}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp unknown tx type; %w", err)
		}
		return nil, nil
	}
}

func slpGenesisObjects(info parse.OpReturn) ([]db.Object, error) {
	const ExpectedPushDataCount = 10
	if len(info.PushData) < ExpectedPushDataCount {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid slp genesis, incorrect push data (%d), expected %d", len(info.PushData), ExpectedPushDataCount),
		}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp genesis incorrect push data; %w", err)
		}
		return nil, nil
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
	var objects = []db.Object{genesis}
	output, err := slpOutputObject(info, genesis.TxHash, memo.SlpMintTokenIndex, jutil.GetUint64(info.PushData[9]))
	if err != nil {
		return nil, fmt.Errorf("error building slp output for genesis; %w", err)
	}
	if output != nil {
		objects = append(objects, output)
	}
	baton, err := slpBatonObject(info, genesis.TxHash, genesis.BatonIndex)
	if err != nil {
		return nil, fmt.Errorf("error building slp baton for genesis; %w", err)
	}
	if baton != nil {
		objects = append(objects, baton)
	}
	return objects, nil
}

func slpMintObjects(info parse.OpReturn) ([]db.Object, error) {
	const ExpectedPushDataCount = 6
	if len(info.PushData) < ExpectedPushDataCount {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid slp mint, incorrect push data (%d), expected %d", len(info.PushData), ExpectedPushDataCount),
		}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp mint incorrect push data; %w", err)
		}
		return nil, nil
	}
	tokenHash, err := chainhash.NewHash(jutil.ByteReverse(info.PushData[3]))
	if err != nil {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid token hash for slp mint (%x)", info.PushData[3]),
		}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp mint invalid token hash; %w", err)
		}
		return nil, nil
	}
	var mint = &slp.Mint{
		TxHash:     info.TxHash,
		TokenHash:  *tokenHash,
		TokenType:  uint8(jutil.GetUint64(info.PushData[1])),
		BatonIndex: uint32(jutil.GetUint64(info.PushData[4])),
		Quantity:   jutil.GetUint64(info.PushData[5]),
	}
	var objects = []db.Object{mint}
	output, err := slpOutputObject(info, mint.TokenHash, memo.SlpMintTokenIndex, mint.Quantity)
	if err != nil {
		return nil, fmt.Errorf("error building slp output for mint; %w", err)
	}
	if output != nil {
		objects = append(objects, output)
	}
	baton, err := slpBatonObject(info, mint.TokenHash, mint.BatonIndex)
	if err != nil {
		return nil, fmt.Errorf("error building slp baton for mint; %w", err)
	}
	if baton != nil {
		objects = append(objects, baton)
	}
	return objects, nil
}

func slpSendObjects(info parse.OpReturn) ([]db.Object, error) {
	const ExpectedPushDataCount = 5
	if len(info.PushData) < ExpectedPushDataCount {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error: fmt.Sprintf("invalid slp send, incorrect push data (%d), expected %d",
				len(info.PushData), ExpectedPushDataCount)}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp send incorrect push data; %w", err)
		}
		return nil, nil
	}
	tokenHash, err := chainhash.NewHash(jutil.ByteReverse(info.PushData[3]))
	if err != nil {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid token hash for slp send (%x)", info.PushData[3]),
		}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp send invalid token hash; %w", err)
		}
		return nil, nil
	}
	var send = &slp.Send{
		TxHash:    info.TxHash,
		TokenHash: *tokenHash,
		TokenType: uint8(jutil.GetUint64(info.PushData[1])),
	}
	var objects = []db.Object{send}
	for i := 4; i < len(info.PushData); i++ {
		var index = uint32(i - 3)
		var quantity = jutil.GetUint64(info.PushData[i])
		if quantity == 0 {
			continue
		}
		output, err := slpOutputObject(info, send.TokenHash, index, quantity)
		if err != nil {
			return nil, fmt.Errorf("error building slp output for send; %w", err)
		}
		if output != nil {
			objects = append(objects, output)
		}
	}
	return objects, nil
}

func slpOutputObject(info parse.OpReturn, tokenHash [32]byte, index uint32, quantity uint64) (db.Object, error) {
	if quantity == 0 {
		return nil, nil
	}
	if len(info.Outputs) <= int(index) {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error: fmt.Sprintf("invalid slp output, index out of range (len: %d, index: %d)",
				len(info.Outputs), index)}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp output index out of range; %w", err)
		}
		return nil, nil
	}
	return &slp.Output{
		TxHash:    info.TxHash,
		Index:     index,
		TokenHash: tokenHash,
		Quantity:  quantity,
	}, nil
}

func slpBatonObject(info parse.OpReturn, tokenHash [32]byte, index uint32) (db.Object, error) {
	if len(info.Outputs) <= int(index) {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error: fmt.Sprintf("invalid slp baton, index out of range (len: %d, index: %d)",
				len(info.Outputs), index)}); err != nil {
			return nil, fmt.Errorf("error saving process error for slp baton index out of range; %w", err)
		}
		return nil, nil
	}
	return &slp.Baton{
		TxHash:    info.TxHash,
		Index:     index,
		TokenHash: tokenHash,
	}, nil
}
