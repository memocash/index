package op_return

import (
	"bytes"
	"testing"

	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// fakeP2pkhUnlock builds a P2PKH-shaped unlock script (71-byte signature push
// + 33-byte pubkey push); the pubkey byte determines the derived address.
func fakeP2pkhUnlock(pubKeyByte byte) []byte {
	unlock := []byte{71}
	unlock = append(unlock, bytes.Repeat([]byte{0x30}, 71)...)
	unlock = append(unlock, 33)
	unlock = append(unlock, bytes.Repeat([]byte{pubKeyByte}, 33)...)
	return unlock
}

func senderInfo(unlockScripts ...[]byte) parse.OpReturn {
	var info parse.OpReturn
	for _, unlockScript := range unlockScripts {
		info.Inputs = append(info.Inputs, &wire.TxIn{SignatureScript: unlockScript})
	}
	return info
}

// The memo sender is the FIRST input whose unlock script yields an address
// (the memo.cash site rule) — not the last, which the old saver loop
// effectively selected.
func TestGetSenderAddrUsesFirstParseableInput(t *testing.T) {
	first := fakeP2pkhUnlock(1)
	second := fakeP2pkhUnlock(2)
	want, err := wallet.GetAddrFromUnlockScript(first)
	if err != nil {
		t.Fatalf("parse first unlock script: %v", err)
	}
	last, err := wallet.GetAddrFromUnlockScript(second)
	if err != nil {
		t.Fatalf("parse second unlock script: %v", err)
	}
	if *want == *last {
		t.Fatal("fixture unlock scripts must derive distinct addresses")
	}
	addr, err := getSenderAddr(senderInfo(first, second))
	if err != nil {
		t.Fatalf("get sender addr: %v", err)
	}
	if addr == nil {
		t.Fatal("sender addr = nil, want first input's address")
	}
	if *addr != *want {
		t.Fatalf("sender addr = %s, want first input's address %s", addr, want)
	}
}

func TestGetSenderAddrSkipsUnparseableInput(t *testing.T) {
	parseable := fakeP2pkhUnlock(3)
	want, err := wallet.GetAddrFromUnlockScript(parseable)
	if err != nil {
		t.Fatalf("parse unlock script: %v", err)
	}
	addr, err := getSenderAddr(senderInfo(nil, parseable))
	if err != nil {
		t.Fatalf("get sender addr: %v", err)
	}
	if addr == nil {
		t.Fatal("sender addr = nil, want second input's address")
	}
	if *addr != *want {
		t.Fatalf("sender addr = %s, want second input's address %s", addr, want)
	}
}
