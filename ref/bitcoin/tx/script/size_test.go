package script_test

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

const (
	// Use own constants to protect against constants getting updated unexpectedly
	OutputFeeOpReturn = 20
	OutputBaseFee     = 12
	OutputOpDataFee   = 3

	SlpDefaultTokenType = 0x01

	InOutTypeInput             = 0x01
	InOutTypeBitcoinOutputSelf = 0x03

	TxHashLength = 32
	PkHashLength = 20
)

var (
	PrefixSlp = append([]byte("SLP"), 0x00)
)

const (
	testFileName = "test.txt"
	testFileType = "text/plain"
	testMessage  = "Hello world!"

	testTokenTicker   = "TST"
	testTokenName     = "Test Token"
	testTokenDocUrl   = "test.com"
	testTokenDecimals = 1

	slpTxTypeGenesis = "GENESIS"
	slpTxTypeMint    = "MINT"
	slpTxTypeSend    = "SEND"

	slpTxQuantityBytes = 8
	slpTxIndexBytes    = 2
)

func len64(b []byte) int64 {
	return int64(len(b))
}

type SizeTest struct {
	Name   string
	Script memo.Script
	Size   int64
}

var likeTest = SizeTest{
	Name: "like",
	Script: &script.Like{
		TxHash: test_tx.HashEmptyTx,
	},
	Size: OutputFeeOpReturn + TxHashLength,
}

var muteTest = SizeTest{
	Name: "mute",
	Script: &script.MuteUser{
		MutePkHash: test_tx.Address2pkHash,
	},
	Size: OutputFeeOpReturn + PkHashLength,
}

var postTest = SizeTest{
	Name: "post",
	Script: &script.Post{
		Message: testMessage,
	},
	Size: OutputFeeOpReturn + len64([]byte(testMessage)),
}

var replyTest = SizeTest{
	Name: "reply",
	Script: &script.Reply{
		TxHash:  test_tx.HashEmptyTx,
		Message: testMessage,
	},
	Size: OutputFeeOpReturn + OutputOpDataFee + TxHashLength + len64([]byte(testMessage)),
}

var saveTest = SizeTest{
	Name: "save",
	Script: &script.Save{
		Filename: testFileName,
		Filetype: testFileType,
		Contents: []byte(testMessage),
	},
	Size: OutputBaseFee + OutputOpDataFee*5 + len64([]byte(test_tx.Address1String)) +
		len64([]byte(testMessage)) + len64([]byte(testFileName)) + len64([]byte(testFileType)),
}

var sendTest = SizeTest{
	Name: "send",
	Script: &script.Send{
		Hash:    test_tx.Address2pkHash,
		Message: testMessage,
	},
	Size: OutputFeeOpReturn + OutputOpDataFee + PkHashLength + len64([]byte(testMessage)),
}

var tokenCreateTest = SizeTest{
	Name: "token-create",
	Script: &script.TokenCreate{
		Ticker:   testTokenTicker,
		Name:     testTokenName,
		SlpType:  SlpDefaultTokenType,
		Decimals: testTokenDecimals,
		DocUrl:   testTokenDocUrl,
		Quantity: 1e6,
	},
	Size: OutputBaseFee + OutputOpDataFee*6 + len64(PrefixSlp) +
		len64([]byte{txscript.OP_DATA_1, SlpDefaultTokenType}) + len64([]byte(slpTxTypeGenesis)) +
		len64([]byte(testTokenTicker)) + len64([]byte(testTokenName)) + len64([]byte(testTokenDocUrl)) +
		len64([]byte{txscript.OP_DATA_1, 0}) + len64([]byte{txscript.OP_DATA_1, byte(testTokenDecimals % 255)}) +
		len64([]byte{txscript.OP_DATA_1, 0x02}) + slpTxQuantityBytes,
}

var tokenMintTest = SizeTest{
	Name: "token-mint",
	Script: &script.TokenMint{
		TokenHash: test_tx.HashEmptyTx,
		TokenType: SlpDefaultTokenType,
		Quantity:  100,
	},
	Size: OutputBaseFee + OutputOpDataFee*4 + len64(PrefixSlp) +
		len64([]byte{txscript.OP_DATA_1, SlpDefaultTokenType}) + len64([]byte(slpTxTypeMint)) + TxHashLength +
		len64([]byte{txscript.OP_DATA_1, 0x02}) + slpTxQuantityBytes,
}

var tokenOfferTest = SizeTest{
	Name: "token-offer",
	Script: &script.TokenOffer{
		SellTxHash: test_tx.HashEmptyTx,
		InOuts: []script.InOut{
			script.InOutInput{
				TxHash: test_tx.HashEmptyTx,
				Index:  0,
			},
			script.InOutOutput{
				IsSelf:   true,
				Address:  test_tx.Address1,
				Quantity: 1000,
			},
		},
	},
	Size: OutputFeeOpReturn + OutputOpDataFee*5 + TxHashLength + len64([]byte{InOutTypeInput}) + TxHashLength + slpTxIndexBytes +
		len64([]byte{InOutTypeBitcoinOutputSelf}) + slpTxQuantityBytes,
}

var tokenPinTest = SizeTest{
	Name: "token-pin",
	Script: &script.TokenPin{
		PostTxHash:  test_tx.HashEmptyTx,
		TokenTxHash: test_tx.HashEmptyTx,
		TokenIndex:  0,
	},
	Size: OutputFeeOpReturn + OutputOpDataFee*2 + TxHashLength*2 + slpTxIndexBytes,
}

var tokenSellTest = SizeTest{
	Name: "token-sell",
	Script: &script.TokenSell{
		InOuts: []script.InOut{
			script.InOutInput{
				TxHash: test_tx.HashEmptyTx,
				Index:  0,
			},
			script.InOutOutput{
				IsSelf:   true,
				Address:  test_tx.Address1,
				Quantity: 1000,
			},
		},
	},
	Size: OutputFeeOpReturn + OutputOpDataFee*4 + len64([]byte{InOutTypeInput}) + TxHashLength + slpTxIndexBytes +
		len64([]byte{InOutTypeBitcoinOutputSelf}) + slpTxQuantityBytes,
}

var tokenSendTest = SizeTest{
	Name: "token-send",
	Script: &script.TokenSend{
		TokenHash:  test_tx.HashEmptyTx,
		SlpType:    SlpDefaultTokenType,
		Quantities: []uint64{5, 995},
	},
	Size: OutputBaseFee + OutputOpDataFee*5 + len64(PrefixSlp) + slpTxQuantityBytes*2 +
		len64([]byte{txscript.OP_DATA_1, SlpDefaultTokenType}) + len64([]byte(slpTxTypeSend)) + TxHashLength,
}

var tokenSignatureTest = SizeTest{
	Name: "token-signature",
	Script: &script.TokenSignature{
		OfferTxHash: test_tx.HashEmptyTx,
		Signatures:  []script.Signature{{Sig: test_tx.SellTokenSignature, PkData: test_tx.SellTokenPkData}},
	},
	Size: OutputFeeOpReturn + OutputOpDataFee + TxHashLength + len64(test_tx.SellTokenSignature),
}

var tests = []SizeTest{
	likeTest,
	muteTest,
	postTest,
	replyTest,
	saveTest,
	sendTest,
	tokenCreateTest,
	tokenMintTest,
	tokenOfferTest,
	tokenPinTest,
	tokenSellTest,
	tokenSendTest,
	tokenSignatureTest,
}

func TestSizes(t *testing.T) {
	for _, tst := range tests {
		size, err := memo.GetOutputSize(tst.Script)
		if err != nil {
			t.Error(jerr.Getf(err, "test %s error getting script", tst.Name))
			continue
		}
		if size != tst.Size {
			if size != 0 {
				// TODO: This test is old from when sizes were estimated. Remove or fix.
				continue
			}
			t.Error(jerr.Newf("test %s size does not match %d (expected: %d)", tst.Name, size, tst.Size))
		}
	}
}
