package wallet

import (
	"bytes"
)

type KeyRing struct {
	Keys []PrivateKey
}

func (k KeyRing) GetKey(pkHash []byte) PrivateKey {
	for _, key := range k.Keys {
		if bytes.Equal(key.GetPkHash(), pkHash) {
			return key
		}
	}
	return PrivateKey{}
}

func GetSingleKeyRing(privateKey PrivateKey) KeyRing {
	return KeyRing{Keys: []PrivateKey{privateKey}}
}
