package wallet

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/btcd/btcec"
	"github.com/jchavannes/btcutil"
	"github.com/jchavannes/btcutil/base58"
	"github.com/jchavannes/jgo/jerr"
)

func GetPrivateKey(secretHex string) PrivateKey {
	secret, _ := hex.DecodeString(secretHex)
	return PrivateKey{
		Secret: secret,
	}
}

func GeneratePrivateKey() PrivateKey {
	b := make([]byte, 32)
	rand.Read(b)
	return PrivateKey{
		Secret: b,
	}
}

func ImportPrivateKey(wifString string) (PrivateKey, error) {
	wif, err := btcutil.DecodeWIF(wifString)
	if err != nil {
		return PrivateKey{}, jerr.Get("error creating wif", err)
	}
	return PrivateKey{
		Secret: wif.PrivKey.Serialize(),
	}, nil
}

func ImportPrivateKeyNew(wifString string) PrivateKey {
	pk, _ := ImportPrivateKey(wifString)
	return pk
}

type PrivateKey struct {
	Secret []byte
}

func (k PrivateKey) IsSet() bool {
	return len(k.Secret) > 0
}

func (k PrivateKey) GetBinaryString() string {
	var binaryKey string
	for _, n := range k.Secret {
		binaryKey += fmt.Sprintf("%b", n)
	}
	return binaryKey
}

func (k PrivateKey) GetBase58() string {
	return base58.CheckEncode(k.Secret, 128)
}

func (k PrivateKey) GetBase58Compressed() string {
	return base58.CheckEncode(append(k.Secret, 0x01), 128)
}

func (k PrivateKey) GetHex() string {
	return hex.EncodeToString(k.Secret)
}

func (k PrivateKey) GetHexCompressed() string {
	return hex.EncodeToString(append(k.Secret, 0x01))
}

func (k PrivateKey) GetPublicKey() PublicKey {
	_, pub := btcec.PrivKeyFromBytes(btcec.S256(), k.Secret)
	return PublicKey{
		publicKey: pub,
	}
}

func (k PrivateKey) GetPkHash() []byte {
	return k.GetPublicKey().GetPkHash()
}

func (k PrivateKey) GetAddress() Address {
	return k.GetPublicKey().GetAddress()
}

func (k PrivateKey) GetAddr() Addr {
	return *GetAddrFromPkHash(k.GetPkHash())
}

func (k PrivateKey) GetBtcEcPrivateKey() *btcec.PrivateKey {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), k.Secret)
	return priv
}
