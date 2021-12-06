package wallet_test

import (
	"github.com/jchavannes/go-mnemonic/bip39"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_wallet"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

const (
	Mnemonic       = "install bargain aunt notice picnic syrup moral autumn april helmet oil assume"
	ChildAddress0  = "1GvvJq5KpqgvJ2rpScJhKrVPYhiARs9AaT"
	ChildAddress1  = "1826k8BxGhd7S1nukew9tVE4Xgard4qo1n"
	ChildAddress2  = "1AMDYpPV1aYS6NmaiLstanws4mXeXED9hs"
	ChildAddress3  = "1JJhfR2fD3mmxipTXxdtvehBiep1WNzM9q"
	ChildAddress4  = "1GDCrKWYE8TA5ie4YrgJb2xzyAL97RhtjK"
	ChildAddress5  = "1MCgBDVXTwfEKYtu2PtPHBif5BpthvBrHJ"
	ChildAddress6  = "189a3sScpKMQRgbUfJzbHAtEEhW6KnPsBu"
	ChildAddress7  = "1AVZBuAFUdeCPzPXDBmzutuM5UvjuV4Y5Y"
	ChildAddress8  = "1MFWkYMWxJYsAwBXrQnJ8Pd8dwEiNvRuj4"
	ChildAddress9  = "112pUgn7wocPtXiw7U8wJ1TW73tpdoQDFA"
	ChangeAddress0 = "1NJsTohbfrqtzUh2N3bviSnrnNvfscyFkY"
)

type PathTest struct {
	Path    string
	Address string
}

var pathTests = []PathTest{{
	Path:    test_wallet.BtcPathAddress0,
	Address: ChildAddress0,
}, {
	Path:    test_wallet.BtcPathAddress1,
	Address: ChildAddress1,
}, {
	Path:    test_wallet.BtcPathChange0,
	Address: ChangeAddress0,
}}

func TestMnemonic(t *testing.T) {
	bip39mnemonic, err := bip39.NewMnemonicFromSentence(Mnemonic, "")
	if err != nil {
		t.Error(jerr.Get("error getting mnemonic", err))
		return
	}
	var mnemonic = wallet.Mnemonic{Mnemonic: *bip39mnemonic}
	for _, pathTest := range pathTests {
		child, err := mnemonic.GetPath(pathTest.Path)
		if err != nil {
			t.Error(jerr.Get("error getting child key from mnemonic", err))
			return
		}
		childAddress := child.GetPublicKey().GetAddress().GetEncoded()
		if childAddress != pathTest.Address {
			t.Error(jerr.Newf("child address %s does not match expected %s", childAddress, pathTest.Address))
		} else if testing.Verbose() {
			jlog.Logf("childAddress: %s, expected: %s\n", childAddress, pathTest.Address)
		}
	}
}
