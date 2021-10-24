package test_tx

import (
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

func GetHexBytes(hash string) []byte {
	h, _ := hex.DecodeString(hash)
	return h
}

func GetHashBytes(hash string) []byte {
	h, _ := chainhash.NewHashFromStr(hash)
	return h.CloneBytes()
}

func GetAddressPkHash(address string) []byte {
	return wallet.GetAddressFromString(address).GetPkHash()
}

func GetPrivateKey(wif string) wallet.PrivateKey {
	key, _ := wallet.ImportPrivateKey(wif)
	return key
}

func GetAddress(address string) wallet.Address {
	return wallet.GetAddressFromString(address)
}

func GetBlockHeader(raw string) wire.BlockHeader {
	r, _ := hex.DecodeString(raw)
	header, _ := memo.GetBlockHeaderFromRaw(r)
	return *header
}
