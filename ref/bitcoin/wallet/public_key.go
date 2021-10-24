package wallet

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/jchavannes/jgo/jerr"
)

func GetPublicKey(pkBytes []byte) (PublicKey, error) {
	pubKey, err := btcec.ParsePubKey(pkBytes, btcec.S256())
	if err != nil {
		return PublicKey{}, jerr.Get("error parsing pub key", err)
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
