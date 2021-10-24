package wallet

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	chainhash2 "github.com/jchavannes/btcd/chaincfg/chainhash"
)

func ConvertChainHashToBCH(hash *chainhash.Hash) *chainhash2.Hash {
	newHash, _ := chainhash2.NewHash(hash.CloneBytes())
	return newHash
}

func ConvertChainHashToBTC(hash *chainhash2.Hash) *chainhash.Hash {
	newHash, _ := chainhash.NewHash(hash.CloneBytes())
	return newHash
}
