package wallet_test

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

const (
	MnemonicEntropy            = "0c1e24e5917779d297e14d45f14e1a1a"
	MnemonicWords              = "army van defense carry jealous true garbage claim echo media make crunch"
	MnemonicPassphrase         = "SuperDuperSecret"
	MnemonicSeedNoPassphrase   = "5b56c417303faa3fcba7e57400e120a0ca83ec5a4fc9ffba757fbe63fbd77a89a1a3be4c67196f57c39a88b76373733891bfaba16ed27a813ceed498804c0570"
	MnemonicSeedWithPassphrase = "3b5df16df2157104cfdd22830162a5e170c0161653e3afe6c88defeefb0818c793dbb28ab3ab091897d0715861dc8a18358f80b79d49acf64142ae57037d1d54"
)

const (
	Mnemonic256Entropy = "2041546864449caff939d32d574753fe684d3c947c3346713dd8423e74abcf8c"
	Mnemonic256Words   = "cake apple borrow silk endorse fitness top denial coil riot stay wolf luggage oxygen faint major edit measure invite love trap field dilemma oblige"
	Mnemonic256Seed    = "3269bce2674acbd188d4f120072b13b088a0ecf87c6e4cae41657a0bb78f5315b33b3a04356e53d062e55f1e0deaa082df8d487381379df848a6ad7e98798404"
)

func TestMnemonicNoPassphrase(t *testing.T) {
	noPassphraseMnemonic, err := wallet.GetWallet(MnemonicWords, "")
	if err != nil {
		t.Error(jerr.Get("error getting wallet", err))
		t.FailNow()
	}
	if noPassphraseMnemonic.GetEntropy() != MnemonicEntropy {
		t.Error(jerr.New("entropy does not match"))
		t.FailNow()
	}
	seed, err := noPassphraseMnemonic.GetSeed()
	if err != nil {
		t.Error(jerr.Get("error getting seed", err))
		t.FailNow()
	}
	if seed != MnemonicSeedNoPassphrase {
		t.Error(jerr.New(fmt.Sprintf("seed (%s) does not match expected (%s)", seed, MnemonicSeedNoPassphrase)))
		t.FailNow()
	}
	fmt.Printf("- Seed without passphrase matches.\n  Seed:     %s\n  Expected: %s\n", seed, MnemonicSeedNoPassphrase)
}

func TestMnemonicWithPassphrase(t *testing.T) {
	withPassphraseMnemonic, err := wallet.GetWallet(MnemonicWords, MnemonicPassphrase)
	if err != nil {
		t.Error(jerr.Get("error getting wallet", err))
		t.FailNow()
	}
	if withPassphraseMnemonic.GetEntropy() != MnemonicEntropy {
		t.Error(jerr.New("entropy does not match"))
		t.FailNow()
	}
	seed, err := withPassphraseMnemonic.GetSeed()
	if err != nil {
		t.Error(jerr.Get("error getting seed", err))
		t.FailNow()
	}
	if seed != MnemonicSeedWithPassphrase {
		t.Error(jerr.New(fmt.Sprintf("seed (%s) does not match expected (%s)", seed, MnemonicSeedWithPassphrase)))
		t.FailNow()
	}
	fmt.Printf("- Seed with passphrase matches.\n  Seed:     %s\n  Expected: %s\n", seed, MnemonicSeedWithPassphrase)
}

func TestMnemonic256(t *testing.T) {
	mnemonic256, err := wallet.GetWallet(Mnemonic256Words, "")
	if err != nil {
		t.Error(jerr.Get("error getting wallet", err))
		t.FailNow()
	}
	if mnemonic256.GetEntropy() != Mnemonic256Entropy {
		t.Error(jerr.New("entropy does not match"))
		t.FailNow()
	}
	seed, err := mnemonic256.GetSeed()
	if err != nil {
		t.Error(jerr.Get("error getting seed", err))
		t.FailNow()
	}
	if seed != Mnemonic256Seed {
		t.Error(jerr.New(fmt.Sprintf("seed (%s) does not match expected (%s)", seed, Mnemonic256Seed)))
		t.FailNow()
	}
	fmt.Printf("- Seed 256 matches.\n  Seed:     %s\n  Expected: %s\n", seed, Mnemonic256Seed)
}
