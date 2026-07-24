package op_return

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func TestMemoSendHandlerRegistered(t *testing.T) {
	pkScript, err := (script.Send{
		Hash:    bytes.Repeat([]byte{1}, memo.PkHashLength),
		Message: "send message",
	}).Get()
	if err != nil {
		t.Fatalf("build send script: %v", err)
	}
	handlers, err := GetHandlers()
	if err != nil {
		t.Fatalf("get handlers: %v", err)
	}
	for _, handler := range handlers {
		if handler == memoSendHandler && handler.CanHandle(pkScript) {
			return
		}
	}
	t.Fatal("Memo Send handler is not registered")
}

func TestGetMemoSendMessage(t *testing.T) {
	pkScript, err := (script.Send{
		Hash:    bytes.Repeat([]byte{1}, memo.PkHashLength),
		Message: "send message",
	}).Get()
	if err != nil {
		t.Fatalf("build send script: %v", err)
	}
	pushData, err := txscript.PushedData(pkScript)
	if err != nil {
		t.Fatalf("decode send script: %v", err)
	}
	message, err := getMemoSendMessage(pushData)
	if err != nil {
		t.Fatalf("get send message: %v", err)
	}
	if message != "send message" {
		t.Fatalf("message = %q, want %q", message, "send message")
	}
}

func TestGetMemoSendMessageRejectsInvalidRecipient(t *testing.T) {
	_, err := getMemoSendMessage([][]byte{
		memo.PrefixSendMoney,
		{1},
		[]byte("send message"),
	})
	if err == nil {
		t.Fatal("getMemoSendMessage error = nil, want invalid recipient error")
	}
}

func TestGetMemoSendRecipientMissing(t *testing.T) {
	hash := bytes.Repeat([]byte{1}, memo.PkHashLength)
	otherOutput := mustDecodeHex(t, "76a914020202020202020202020202020202020202020288ac")
	if _, err := getMemoSendRecipient(hash, []*wire.TxOut{
		wire.NewTxOut(1, otherOutput),
	}); err == nil {
		t.Fatal("getMemoSendRecipient error = nil, want missing recipient error")
	}
}

func TestReportedMemoSendFixture(t *testing.T) {
	// Output and input scripts from
	// 46f1cea809d9ed63751cad305ecb5c9d01df36173db898effef216e1b588989c.
	opReturn := mustDecodeHex(t, "6a026d2414910a0201437732db6cbb13083adbb62c7ab960044cbf486920667265657472616465722e20436f756c6420796f75206769766520686967682d6c6576656c205554584f2f73637269707420666565646261636b206f6e205465796f6c6961205845432f52554e45206265666f726520616e79206361706974616c2061736b3f204e6f20656e646f7273656d656e74207265717565737465642e2068747470733a2f2f6769746c61622e636f6d2f586f6c6f73524d5a2f7465796f6c69612d7865632d72756e652d7265766965772d7061636b616765")
	recipientOutput := mustDecodeHex(t, "76a914910a0201437732db6cbb13083adbb62c7ab9600488ac")
	input := mustDecodeHex(t, "483045022100e3596623d663dea77c7270d63887593ce5a1e419361b26bd2175c2f6ea308af802202a38d39cd1493491411ebbff0c0e27cfe4445becb32befacd37109c9a5d6803e4121023d808dd01c2fcae1c3aa769026581ebdb677b603117e8079645c961d0e966868")

	pushData, err := txscript.PushedData(opReturn)
	if err != nil {
		t.Fatalf("decode reported send script: %v", err)
	}
	message, err := getMemoSendMessage(pushData)
	if err != nil {
		t.Fatalf("get reported send message: %v", err)
	}
	wantMessage := "Hi freetrader. Could you give high-level UTXO/script feedback on Teyolia XEC/RUNE before any capital ask? No endorsement requested. https://gitlab.com/XolosRMZ/teyolia-xec-rune-review-package"
	if message != wantMessage {
		t.Fatalf("message = %q, want %q", message, wantMessage)
	}
	recipient, err := getMemoSendRecipient(pushData[1], []*wire.TxOut{
		wire.NewTxOut(33333, recipientOutput),
	})
	if err != nil {
		t.Fatalf("get reported send recipient: %v", err)
	}
	if got := recipient.String(); got != "1EDtyAxn8zh9qBdWxdCyhY7AFzLcHvFrtE" {
		t.Fatalf("recipient = %q, want reported recipient", got)
	}
	author, err := wallet.GetAddrFromUnlockScript(input)
	if err != nil {
		t.Fatalf("get reported send author: %v", err)
	}
	if got := author.String(); got != "1Kq5hxjgyzTow9KTMzBJDuSeW3S2eHpXhx" {
		t.Fatalf("author = %q, want reported signer", got)
	}
	if *author == recipient {
		t.Fatal("reported send author unexpectedly equals recipient")
	}
}

func mustDecodeHex(t *testing.T, value string) []byte {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decode fixture hex: %v", err)
	}
	return decoded
}
