package script

import (
	"encoding/binary"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type TokenSend struct {
	TokenHash  []byte
	SlpType    byte
	Quantities []uint64
}

func (t TokenSend) GetTotalQuantity() uint64 {
	var total uint64
	for _, quantity := range t.Quantities {
		total += quantity
	}
	return total
}

func (t TokenSend) Get() ([]byte, error) {
	if t.SlpType == 0 {
		return nil, jerr.New("type not set")
	}
	script := memo.GetBaseOpReturn().
		AddData(memo.PrefixSlp).
		AddOps([]byte{txscript.OP_DATA_1, t.SlpType}).
		AddData([]byte(memo.SlpTxTypeSend)).
		AddData(jutil.ByteReverse(t.TokenHash))
	quantities := t.Quantities
	for i := len(quantities) - 1; i > 0; i-- {
		if quantities[i] == 0 {
			quantities = quantities[:i]
		} else {
			break
		}
	}
	for _, quantity := range quantities {
		var quantityBytes = make([]byte, memo.Int8Size)
		binary.BigEndian.PutUint64(quantityBytes, quantity)
		script = script.AddData(quantityBytes)
	}
	pkScript, err := script.Script()
	if err != nil {
		return nil, jerr.Get("error building script", err)
	}
	return pkScript, nil
}

func (t TokenSend) Type() memo.OutputType {
	return memo.OutputTypeTokenSend
}
