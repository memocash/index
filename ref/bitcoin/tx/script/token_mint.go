package script

import (
	"encoding/binary"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type TokenMint struct {
	TokenHash []byte
	TokenType byte
	Quantity  uint64
}

func (t TokenMint) Get() ([]byte, error) {
	if t.TokenType == 0 {
		return nil, jerr.New("token type not set")
	}
	var quantityBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(quantityBytes, t.Quantity)
	var batonVOut = []byte{txscript.OP_DATA_1, 0x02}
	var script = memo.GetBaseOpReturn().
		AddData(memo.PrefixSlp).
		AddOps([]byte{txscript.OP_DATA_1, t.TokenType}).
		AddData([]byte(memo.SlpTxTypeMint)).
		AddData(jutil.ByteReverse(t.TokenHash)).
		AddOps(batonVOut).
		AddData(quantityBytes)
	pkScript, err := script.Script()
	if err != nil {
		return nil, jerr.Get("error building script", err)
	}
	return pkScript, nil
}

func (t TokenMint) Type() memo.OutputType {
	return memo.OutputTypeTokenMint
}
