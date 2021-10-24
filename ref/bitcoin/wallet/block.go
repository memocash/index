package wallet

import "github.com/btcsuite/btcd/chaincfg/chainhash"

type Block struct {
	Hash       *chainhash.Hash
	MerkleRoot *chainhash.Hash
}

var _genesisBlock *Block
var _firstBlock *Block

func GetGenesisBlock() Block {
	if _genesisBlock == nil {
		hash, _ := chainhash.NewHashFromStr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
		merkleRoot, _ := chainhash.NewHashFromStr("4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b")
		_genesisBlock = &Block{
			Hash:       hash,
			MerkleRoot: merkleRoot,
		}
	}
	return *_genesisBlock
}

func GetFirstBlock() Block {
	if _firstBlock == nil {
		hash, _ := chainhash.NewHashFromStr("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048")
		merkleRoot, _ := chainhash.NewHashFromStr("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098")
		_firstBlock = &Block{
			Hash:       hash,
			MerkleRoot: merkleRoot,
		}
	}
	return *_firstBlock
}
