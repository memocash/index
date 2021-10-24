package wallet

func ConvertSatoshisToBch(satoshis uint64) float64 {
	return ConvertFloatSatoshisToBch(float64(satoshis))
}

func ConvertFloatSatoshisToBch(satoshis float64) float64 {
	return satoshis * 1e-8
}
