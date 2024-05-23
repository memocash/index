package wallet

import (
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/go-mnemonic/bip39"
)

func GetWallet(mnemonicPhrase string, passphrase string) (Wallet, error) {
	mnemonic, err := bip39.NewMnemonicFromSentence(mnemonicPhrase, passphrase)
	if err != nil {
		return Wallet{}, fmt.Errorf("error getting mnemonic from sentence; %w", err)
	}
	entropyHex, err := mnemonic.GetEntropyStrHex()
	if err != nil {
		return Wallet{}, fmt.Errorf("error getting entropy from mnemonic; %w", err)
	}
	entropy, err := hex.DecodeString(entropyHex)
	if err != nil {
		return Wallet{}, fmt.Errorf("error decoding entropy hex; %w", err)
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
		return "", fmt.Errorf("error getting mnemonic from entropy; %w", err)
	}
	seed, err := mnemonic.GetSeed()
	if err != nil {
		return "", fmt.Errorf("error getting seed from mnemonic; %w", err)
	}
	return seed, nil
}
