package wallet

import (
	"encoding/hex"
	"github.com/jchavannes/go-mnemonic/bip39"
	"github.com/jchavannes/jgo/jerr"
)

func GetWallet(mnemonicPhrase string, passphrase string) (Wallet, error) {
	mnemonic, err := bip39.NewMnemonicFromSentence(mnemonicPhrase, passphrase)
	if err != nil {
		return Wallet{}, jerr.Get("error getting mnemonic from sentence", err)
	}
	entropyHex, err := mnemonic.GetEntropyStrHex()
	if err != nil {
		return Wallet{}, jerr.Get("error getting entropy from mnemonic", err)
	}
	entropy, err := hex.DecodeString(entropyHex)
	if err != nil {
		return Wallet{}, jerr.Get("error decoding entropy hex", err)
	}
	return Wallet{
		Entropy:    entropy,
		Passphrase: passphrase,
	}, nil
}

type Wallet struct {
	Entropy    []byte
	Passphrase string
}

func (w *Wallet) GetEntropy() string {
	return hex.EncodeToString(w.Entropy)
}

func (w *Wallet) GetSeed() (string, error) {
	mnemonic, err := bip39.NewMnemonicFromEntropy(w.Entropy, w.Passphrase)
	if err != nil {
		return "", jerr.Get("error getting mnemonic from entropy", err)
	}
	seed, err := mnemonic.GetSeed()
	if err != nil {
		return "", jerr.Get("error getting seed from mnemonic", err)
	}
	return seed, nil
}
