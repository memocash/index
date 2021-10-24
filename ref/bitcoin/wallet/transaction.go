package wallet

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/bchutil"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
)

func DecodeTx(txHex string) (*wire.MsgTx, error) {
	tx1Byte, _ := hex.DecodeString(txHex)
	tx, err := bchutil.NewTxFromBytes(tx1Byte)
	if err != nil {
		return nil, jerr.Get("error getting tx from bytes", err)
	}
	return tx.MsgTx(), nil
}

func EncodeTx(tx *wire.MsgTx) string {
	writer := new(bytes.Buffer)
	tx.BtcEncode(writer, 1)
	return hex.EncodeToString(writer.Bytes())
}
