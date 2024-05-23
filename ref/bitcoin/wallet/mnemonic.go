package wallet

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg"
	"github.com/jchavannes/btcutil/hdkeychain"
	"github.com/jchavannes/go-mnemonic/bip39"
	"github.com/memocash/index/ref/bitcoin/util"
	"github.com/tyler-smith/go-bip32"
	"strconv"
	"strings"
)

type Mnemonic struct {
	Mnemonic bip39.Mnemonic
	Sentence string
}

func (m Mnemonic) IsSet() bool {
	return len(m.Sentence) > 0
}

// BIP32 / BIP44
func (m *Mnemonic) GetPathExtended(path string) (*hdkeychain.ExtendedKey, error) {
	sentence, err := m.Mnemonic.GetSentence()
	seed := bip39.NewSeed(sentence, "")
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("error getting master key from mnemonic; %w", err)
	}
	newKey := masterKey
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if i == 0 && part == "m" || part == "" {
			continue
		}
		var hardened bool
		if strings.Contains(part, "'") {
			hardened = true
			part = strings.Replace(part, "'", "", -1)
		}
		index, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("error parsing index from part: %s; %w", part, err)
		}
		uIndex := uint32(index)
		if hardened {
			uIndex += bip32.FirstHardenedChild
		}
		newKey, err = newKey.Derive(uIndex)
		if err != nil {
			return nil, fmt.Errorf("error getting child key for index: %d; %w", uIndex, err)
		}
	}
	return newKey, nil
}

func (m *Mnemonic) GetPath(path string) (*PrivateKey, error) {
	newKey, err := m.GetPathExtended(path)
	if err != nil {
		return nil, fmt.Errorf("error getting path extended; %w", err)
	}
	childPrivateKey, err := newKey.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("error getting child private key; %w", err)
	}
	return &PrivateKey{Secret: childPrivateKey.Serialize()}, nil
}

func GetNewMnemonic(mnemonic bip39.Mnemonic) *Mnemonic {
	sentence, _ := mnemonic.GetSentence()
	return &Mnemonic{
		Mnemonic: mnemonic,
		Sentence: sentence,
	}
}

func GenerateMnemonic() (*Mnemonic, error) {
	util.SeedRandom()
	b := make([]byte, 16)
	rand.Read(b)
	code, err := bip39.NewMnemonicFromEntropy(b, "")
	if err != nil {
		return nil, fmt.Errorf("error getting new mnemonic; %w", err)
	}
	return GetNewMnemonic(*code), nil
}

func GetMnemonic(entropy []byte) (*Mnemonic, error) {
	mnemonic, err := bip39.NewMnemonicFromEntropy(entropy, "")
	if err != nil {
		return nil, fmt.Errorf("error getting mnemonic from entropy; %w", err)
	}
	return GetNewMnemonic(*mnemonic), nil
}

var parsingMnemonicError = fmt.Errorf("error getting mnemonic from entropy")

func IsParsingMnemonicError(err error) bool {
	return errors.Is(err, parsingMnemonicError)
}

func GetMnemonicFromString(sentence string) (*Mnemonic, error) {
	mnemonic, err := bip39.NewMnemonicFromSentence(sentence, "")
	if err != nil {
		return nil, fmt.Errorf("%s; %w", err, parsingMnemonicError)
	}
	return GetNewMnemonic(*mnemonic), nil
}
