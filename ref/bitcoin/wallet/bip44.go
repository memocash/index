package wallet

import "fmt"

const (
	Bip44CoinTypeBTC = 0
	Bip44CoinTypeBCH = 145
	Bip44CoinTypeBSV = 236
	Bip44CoinTypeSLP = 245
)

func GetSatoshiPath() string {
	return GetBip44Path(Bip44CoinTypeBTC, 0, false)
}

func GetSLPPath() string {
	return GetBip44Path(Bip44CoinTypeSLP, 0, false)
}

func GetBip44Path(coinType, index uint, change bool) string {
	var changeId uint
	if change {
		changeId = 1
	}
	return fmt.Sprintf("m/44'/%d'/0'/%d/%d", coinType, changeId, index)
}

func GetBip44CoinPath(coinType uint) string {
	return fmt.Sprintf("m/44'/%d'/0'", coinType)
}
