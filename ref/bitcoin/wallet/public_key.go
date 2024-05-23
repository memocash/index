package wallet

import (
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/btcd/btcec"
)

func GetPublicKey(pkBytes []byte) (PublicKey, error) {
	pubKey, err := btcec.ParsePubKey(pkBytes, btcec.S256())
	if err != nil {
		return PublicKey{}, fmt.Errorf("error parsing pub key; %w", err)
	}
	return PublicKey{
		publicKey: pubKey,
	}, nil
}

type PublicKey struct {
	publicKey *btcec.PublicKey
}

func (k PublicKey) GetSerialized() []byte {
	if k.publicKey == nil {
		return []byte{}
	}
	return k.publicKey.SerializeCompressed()
}

func (k PublicKey) GetSerializedString() string {
	return hex.EncodeToString(k.GetSerialized())
}

func (k PublicKey) GetAddress() Address {
	return GetAddress(k.GetSerialized())
}

func (k PublicKey) GetPkHash() []byte {
	return k.GetAddress().GetPkHash()
}

func (k PublicKey) GetBtcEcPubKey() *btcec.PublicKey {
	return k.publicKey
}
