package hs

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func GetTxString(txHash []byte) string {
	ch, err := chainhash.NewHash(txHash)
	if err != nil {
		return ""
	}
	return ch.String()
}

const shortLenUrl = 10

func GetTxHash(hashString string) []byte {
	hash, _ := chainhash.NewHashFromStr(hashString)
	if hash != nil {
		return hash.CloneBytes()
	}
	return nil
}

func GetTxStringShort(txHash []byte) string {
	hashString := GetTxString(txHash)
	if len(hashString) != 64 {
		return hashString
	}
	return jutil.ShortHash(hashString)
}

func GetTxStringShortUrl(txHash []byte) string {
	hashString := GetTxString(txHash)
	if len(hashString) != 64 {
		return hashString
	}
	hashRunes := []rune(hashString)
	return string(hashRunes[:shortLenUrl])
}

func GetAddrString(pkHash []byte) string {
	addr, err := wallet.GetAddressFromPkHashNew(pkHash)
	if err != nil {
		return ""
	}
	return addr.GetEncoded()
}

func GetCashAddrString(pkHash []byte) string {
	addr, err := wallet.GetAddressFromPkHashNew(pkHash)
	if err != nil {
		return ""
	}
	return addr.GetCashAddrString()
}

func GetSLPAddrString(pkHash []byte) string {
	addr, err := wallet.GetAddressFromPkHashNew(pkHash)
	if err != nil {
		return ""
	}
	return addr.GetSlpAddrString()
}

func GetHashIndexString(txHash []byte, index uint32) string {
	ch, err := chainhash.NewHash(txHash)
	if err != nil {
		return ""
	}
	return GetHashIndexWithString(ch.String(), index)
}

func GetHashIndexWithString(txHash string, index uint32) string {
	return fmt.Sprintf("%s:%d", txHash, index)
}

// HashesToSlices converts a slice of 32 byte hashes to a slice of byte slices
func HashesToSlices(txHashes [][32]byte) [][]byte {
	var slices = make([][]byte, len(txHashes))
	for i, txHash := range txHashes {
		slices[i] = txHash[:]
	}
	return slices
}
