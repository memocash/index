package memo

import (
	"github.com/jchavannes/jgo/jfmt"
	"math"
)

type SlpType string

const (
	SlpTxTypeGenesis SlpType = "GENESIS"
	SlpTxTypeMint    SlpType = "MINT"
	SlpTxTypeSend    SlpType = "SEND"
	SlpTxTypeCommit  SlpType = "COMMIT"

	SlpMintTokenIndex = 1
)

const (
	SlpDefaultTokenType  = 0x01
	SlpNftGroupTokenType = 0x81
	SlpNftChildTokenType = 0x41

	SlpNftChildBatonVOut = 0x4c00

	SlpType1        = "Normal"
	SlpTypeNftGroup = "NFT Group"
	SlpTypeNftChild = "NFT Child"
)

func SlpTypeByteString(slpType byte) string {
	return SlpTypeString(uint16(slpType))
}

func SlpTypeString(tokenType uint16) string {
	switch tokenType {
	case SlpDefaultTokenType:
		return SlpType1
	case SlpNftGroupTokenType:
		return SlpTypeNftGroup
	case SlpNftChildTokenType:
		return SlpTypeNftChild
	}
	return ""
}

func GetSlpQuantity(quantity uint64, decimals uint8) float64 {
	return GetDecimalValue(float64(quantity), decimals)
}

func GetSlpQuantityString(quantity uint64, decimals uint8) string {
	if decimals == 0 {
		return jfmt.AddCommasUint(quantity)
	}
	return jfmt.AddCommasFloat(GetSlpQuantity(quantity, decimals), 8, 0)
}

func GetDecimalValue(val float64, decimals uint8) float64 {
	return val / math.Pow10(int(decimals))
}

func GetSlpQuantityInt(quantity float64, decimals int) uint64 {
	return uint64(quantity * math.Pow10(decimals))
}
