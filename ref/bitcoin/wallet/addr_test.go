package wallet_test

import (
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type testAddr struct {
	Address      string
	LockScript   []byte
	UnlockScript []byte
}

const Address1 = "19Y1fAy7n2Qz1JhmRUWmWvDKuKrxL6u4J8"
const Address2 = "3FnvDsVYF5sZ1f3qqEsFHPWqMCjbmhK2bZ"

// c95ef0e4da3ed73017696b9c68c97d820702d2dc94450cfa43bba77b0ffb9ab0
var LockScript1, _ = hex.DecodeString("76a9145d9e7351f7836c0a317b3930e310f3bb01c9f5b688ac")
var UnlockScript1, _ = hex.DecodeString("483045022100b2368b572de7d9dabc0bfe25498452a33d438508402a632a0007dcc8eea60c6602205aff160b1b77c86dd31b2d02c375b38b7e3eb0b0e063c5c92cd7aab757a16004412103b622c21c5fd8266b8d2304a1ae7ae8db384a2104a04dd7c0f62c96dae46698d0")

// dd3a6598226764f660cd93d0a9e69e2e535a13b0298db9d5c271b63b094c6df9
var LockScript2, _ = hex.DecodeString("a9149aaf7f617f829cb4fdf412a6b985540bb59b924587")
var UnlockScript2, _ = hex.DecodeString("285602e80351b2757c00a26900cd02a914c1a97e01877e88c0c66e7c947c7ba06300cc78a269687551")

var tests = []testAddr{{
	Address:      Address1,
	LockScript:   LockScript1,
	UnlockScript: UnlockScript1,
}, {
	Address:      Address2,
	LockScript:   LockScript2,
	UnlockScript: UnlockScript2,
}}

func TestUnlockScriptAddr(t *testing.T) {
	for _, test := range tests {
		addrLock, err := wallet.GetAddrFromLockScript(test.LockScript)
		if err != nil {
			t.Error(jerr.Getf(err, "error getting address from lock script: %s", test.Address))
			continue
		}
		if addrLock.String() != test.Address {
			t.Error(jerr.Newf("address mismatch: %s %s", addrLock.String(), test.Address))
			continue
		}
		/*addrUnlock, err := wallet.GetAddrFromUnlockScript(test.UnlockScript)
		if err != nil {
			t.Error(jerr.Getf(err, "error getting address from unlock script: %s", test.Address))
			continue
		}
		if addrUnlock.String() != test.Address {
			t.Error(jerr.Newf("address mismatch: %s %s", addrUnlock.String(), test.Address))
			continue
		}*/
		jlog.Logf("success address: %s\n", test.Address)
	}
}
