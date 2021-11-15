package test_tx

import (
	"encoding/hex"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

const (
	TxHash1LikeTxString = "6463373fb90c0b69f24beb6954393dec2731556905309703c2ff59859b9360c0"
	TxHash2FundTxString = "f4fe0f674e28043ee6a7bacd087df10d37b32c66aceff1689c523352549e0c29"

	Key1String     = "L4y4WGjmK9pJSqPL34voH4KqQzZn2RnW1tgsGTKuWPkphzNbdVHu"
	Address1String = "1Pzdrdoj2NC25GMWknYn18eHYuvLoZ6dpv"

	Key2String     = "Kz79JZc5eiXAgyo6yThGVJbxbXmAXCQ7awt6LvUKwo1AQuwYPups"
	Address2String = "12DKqE6WyCghwFJe2ce55Go6ECQseDe5qQ"

	Key3String     = "L49uYMVKno5qLAvJKrgfkZpjWdqAfwdSqwz565djym7J1HeRyX8F"
	Address3String = "1JJhfR2fD3mmxipTXxdtvehBiep1WNzM9q"

	Key4String     = "L54iTabaFr3XNKEeV68dCsze8Ri5zmEE6iHNpC9SNHCTcEwgSu4M"
	Address4String = "1GDCrKWYE8TA5ie4YrgJb2xzyAL97RhtjK"

	Key5String     = "Kx2npcZvHCuZarFDsJPVLoyyxGM13LjivoxeWxUGxhjQVUD6nWjB"
	Address5String = "1MCgBDVXTwfEKYtu2PtPHBif5BpthvBrHJ"

	Key6String     = "L1aL8aFWnb3Z24nQQp3Vb83s4v7KKhLYQ4V3RvcDiCZWNhdyxKqL"
	Address6String = "189a3sScpKMQRgbUfJzbHAtEEhW6KnPsBu"

	Key7String     = "KwjQE58reb2MWhe5g5CLVFPpy8nGboKm1S5Dat4AfCbrxuodYGaP"
	Address7String = "1AVZBuAFUdeCPzPXDBmzutuM5UvjuV4Y5Y"

	Key8String     = "L1eeMnKxzvrwK9uT2pKtgdqAj5TFJs3vqddwZGu1z2MHEawobkNT"
	Address8String = "1MFWkYMWxJYsAwBXrQnJ8Pd8dwEiNvRuj4"

	Key9String     = "L2x56HWfN8YWXwkpRmSCCCH2QRhR57bNz4Udio2VMFzDwVj6F6D9"
	Address9String = "112pUgn7wocPtXiw7U8wJ1TW73tpdoQDFA"

	AddressP2sh1String = "3MszuAEhVq6pJyG2rxwGXz65d4fXo7nPhF"
)

const EmptyTxHashString1 = "d21633ba23f70118185227be58a63527675641ad37967e2aa461559f577aec43"
const UnsignedTxTestAddress1LockString = "76A9140DC9316AC4FF22F253D06D65C3770ADD9607096588AC"

// From: c2ebe970899db1a81b0791a8a90fc41251ad8e5fc52c1f9afc1f3fa715a50b9c
const SellTokenSignatureString = "3045022100be3275298b230d8809b9c93326dfac2776f87f283c657a34bc7c65c198c9c95502206b9fb2532a1ebe1be6208d67500f506ce5fede5668b1d23a6d0de89663b8c95fc3"
const SellTokenPkDataString = "03605c2b9b7cc8dc1063be5d7b185fb3c1fd2171bc156f4a30ef1e406789fd6631"

const (
	TestPost        = "Test!"
	TestTokenTicker = "TT"
	TestTokenName   = "Test Token"
	TestMessage     = "Send message"
	TestTopic       = "Test Topic"
	TestTokenDocUrl = "https://token.test"
)

const (
	GenericTxHashString0 = "b158efa8e85ef8283481e000f9fb13b12599a8fa58fce12633093762ebd1cb75"
	GenericTxHashString1 = "ad8b36425e100db1b0bb4677dd447cf08babb493afa0fecced1e9f4d13544ad0"
	GenericTxHashString2 = "e0a9936a36780efa0e50e30cb466e8077c70623cba95a28e3b2125754c98aab7"
	GenericTxHashString3 = "60ab643290d4b0a060eb1b61ba59cd939f82202afa763ac0439fb4ad2d9dd22c"
	GenericTxHashString4 = "6b006731657bb3fe005943759aaca356b364e2f66be32611227e377405927c56"
	GenericTxHashString5 = "813a810f59dabf96e571e43bed884c1f9ae71c4e2740c6c4043e5d5e302fca7c"
	GenericTxHashString6 = "9008dbe3734d33bcf312fe18cc065f39d9c94365befefde46b161cb176176861"
	GenericTxHashString7 = "e27ac7d0e0b9d582716e6d970627cb17c25368c972fc0f433738404f41da1d4c"
	GenericTxHashString8 = "9c956f1f2b8fbbb761f51852a0986eba34e3d14b96deed4b2b27d24907b3522d"
	GenericTxHashString9 = "61a92e42261460904dbcb30ad009a432b4b5205eeaed9a18cb312682303d1c67"
)

const (
	// Really blocks 625306-625310
	Block1Hash = "00000000000000000038a66316b28503ca99d50f184b27cb2152d77ae6a38a12"
	Block2Hash = "0000000000000000019d2aa8f36624c66e05171f17ae4fcd37b8b703d99c410a"
	Block3Hash = "0000000000000000010570c7f6fa7c1e2af996b3d09677be7b1e9f0747087f91"
	Block4Hash = "000000000000000001b2c94746b80003bc710084c500e626a5ec3bc0ebf1437d"
	Block5Hash = "000000000000000001a9e4c5ccbb41304375debb63d0bd61291bcf8e8be1c890"

	Block1Raw = "00000020aa2b15d98f27cbf443f012f3aef5a8d7dcd397d6345692000000000000000000d6e2045acb7ca76b0c804d2e15beab45d0c5c8fca66c58d6379f107f322029beadc4625e2cdd0118b1921204"
)

var (
	UnsignedTxTestAddress1PkScript, _ = hex.DecodeString(UnsignedTxTestAddress1LockString)

	Address1       = GetAddress(Address1String)
	Address1pkHash = GetAddressPkHash(Address1String)
	Address1key    = GetPrivateKey(Key1String)

	Address2       = GetAddress(Address2String)
	Address2pkHash = GetAddressPkHash(Address2String)
	Address2key    = GetPrivateKey(Key2String)

	Address3       = GetAddress(Address3String)
	Address3pkHash = GetAddressPkHash(Address3String)
	Address3key    = GetPrivateKey(Key3String)

	Address4       = GetAddress(Address4String)
	Address4pkHash = GetAddressPkHash(Address4String)
	Address4key    = GetPrivateKey(Key4String)

	Address5       = GetAddress(Address5String)
	Address5pkHash = GetAddressPkHash(Address5String)
	Address5key    = GetPrivateKey(Key5String)

	Address6       = GetAddress(Address6String)
	Address6pkHash = GetAddressPkHash(Address6String)

	Address7       = GetAddress(Address7String)
	Address7pkHash = GetAddressPkHash(Address7String)

	Address8       = GetAddress(Address8String)
	Address8pkHash = GetAddressPkHash(Address8String)

	Address9       = GetAddress(Address9String)
	Address9pkHash = GetAddressPkHash(Address9String)

	AddressP2sh1           = GetAddress(AddressP2sh1String)
	AddressP2sh1scriptHash = GetAddressPkHash(AddressP2sh1String)

	HashEmptyTx = GetHashBytes(EmptyTxHashString1)
	Hash2FundTx = GetHashBytes(TxHash2FundTxString)

	SlpToken1M10 = GetHashBytes("5ce4758425a370a68fe9a644d437b56667fad1ddf9fdf79ddfab784a6c27d466")

	SellTokenSignature = GetHexBytes(SellTokenSignatureString)
	SellTokenPkData    = GetHexBytes(SellTokenPkDataString)

	GenericTxHash0 = GetHashBytes(GenericTxHashString0)
	GenericTxHash1 = GetHashBytes(GenericTxHashString1)
	GenericTxHash2 = GetHashBytes(GenericTxHashString2)
	GenericTxHash3 = GetHashBytes(GenericTxHashString3)
	GenericTxHash4 = GetHashBytes(GenericTxHashString4)
	GenericTxHash5 = GetHashBytes(GenericTxHashString5)
	GenericTxHash6 = GetHashBytes(GenericTxHashString6)
	GenericTxHash7 = GetHashBytes(GenericTxHashString7)
	GenericTxHash8 = GetHashBytes(GenericTxHashString8)
	GenericTxHash9 = GetHashBytes(GenericTxHashString9)

	Block1Header = GetBlockHeader(Block1Raw)
)

func GetAddress1WalletSingle100k() build.Wallet {
	return build.Wallet{
		Getter:  gen.GetWrapper(&TestGetter{UTXOs: []memo.UTXO{Address1InputUtxo100k}}, Address1pkHash),
		Address: Address1,
		KeyRing: wallet.GetSingleKeyRing(Address1key),
	}
}

func GetAddress2WalletEmpty() build.Wallet {
	return build.Wallet{
		Getter:  gen.GetWrapper(&TestGetter{}, Address2pkHash),
		Address: Address2,
		KeyRing: wallet.GetSingleKeyRing(Address2key),
	}
}

func GetWallet(key wallet.PrivateKey, amount int64, index uint32) build.Wallet {
	return build.Wallet{
		Getter: gen.GetWrapper(&TestGetter{UTXOs: []memo.UTXO{{
			Input: memo.TxInput{
				PkHash:       key.GetPkHash(),
				PrevOutHash:  GenericTxHash0,
				PrevOutIndex: index,
				Value:        amount,
			},
		}}}, key.GetPkHash()),
		Address: key.GetAddress(),
		KeyRing: wallet.GetSingleKeyRing(key),
	}
}

var utxoIndex uint32

func ResetUTXOIndex() {
	utxoIndex = 0
}

func GetUTXO(value int64) memo.UTXO {
	indexToUse := utxoIndex
	utxoIndex++
	return memo.UTXO{
		Input: memo.TxInput{
			PkHash:       Address1pkHash,
			PrevOutHash:  HashEmptyTx,
			PrevOutIndex: indexToUse,
			Value:        value,
		},
	}
}

func GetUtxosTestSet1() []memo.UTXO {
	ResetUTXOIndex()
	return []memo.UTXO{
		GetUTXO(123660),
		GetUTXO(1666),
		GetUTXO(699),
		GetUTXO(555),
		GetUTXO(555),
		GetUTXO(555),
	}
}

const UtxosTestSet1MaxSend = 126758

var UtxosSingle25k = []memo.UTXO{{
	Input: memo.TxInput{
		PkScript:    UnsignedTxTestAddress1PkScript,
		PkHash:      Address1pkHash,
		PrevOutHash: HashEmptyTx,
		Value:       25000,
	},
}}

var Address1InputUtxo100k = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        100000,
	},
}

var Address1InputUtxo8k = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        8000,
	},
}

var Address1InputUtxo1k = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        1000,
	},
}

var Address1InputUtxo700 = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        700,
	},
}

var Address1InputUtxo1255 = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        1255,
	},
}

var Address1InputUtxo10070 = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        10070,
	},
}

var Address1InputUtxo861 = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        861,
	},
}

var Address1InputToken = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 1,
		Value:        memo.DustMinimumOutput,
	},
	SlpQuantity: 5,
	SlpToken:    SlpToken1M10,
	SlpType:     memo.SlpTxTypeSend,
}

var Address1InputTokenBaton = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 1,
		Value:        memo.DustMinimumOutput,
	},
	SlpToken: SlpToken1M10,
	SlpType:  memo.SlpTxTypeGenesis,
}

var LikeEmptyPostOutput = memo.Output{
	Script: &script.Like{
		TxHash: HashEmptyTx,
	},
}

var NewPostOutput = memo.Output{
	Script: &script.Post{
		Message: "test",
	},
}

var SetNameOutput = memo.Output{
	Script: &script.SetName{
		Name: "test",
	},
}

var UtxosAddress1twoRegular = []memo.UTXO{{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        2000,
	},
}, {
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 1,
		Value:        memo.DustMinimumOutput,
	},
}}

var UtxosAddress1twoRegularWithToken = []memo.UTXO{{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 0,
		Value:        2000,
	},
}, Address1InputToken}

var Address2InputUtxo100k = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address2pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 2,
		Value:        100000,
	},
}

var Address2Input5Tokens1 = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 1,
		Value:        memo.DustMinimumOutput,
	},
	SlpQuantity: 5,
	SlpToken:    SlpToken1M10,
	SlpType:     memo.SlpTxTypeSend,
}

var Address2Input5Tokens2 = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 2,
		Value:        memo.DustMinimumOutput,
	},
	SlpQuantity: 5,
	SlpToken:    SlpToken1M10,
	SlpType:     memo.SlpTxTypeSend,
}

var Address2Input5Tokens3 = memo.UTXO{
	Input: memo.TxInput{
		PkHash:       Address1pkHash,
		PrevOutHash:  Hash2FundTx,
		PrevOutIndex: 3,
		Value:        memo.DustMinimumOutput,
	},
	SlpQuantity: 5,
	SlpToken:    SlpToken1M10,
	SlpType:     memo.SlpTxTypeSend,
}

var Address2InputsAll3Tokens = []memo.UTXO{
	Address2Input5Tokens1,
	Address2Input5Tokens2,
	Address2Input5Tokens3,
}
