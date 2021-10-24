package memo

func GetMaxSendForUTXOs(utxos []UTXO) int64 {
	var totalValue int64
	var numInputs int
	for _, utxo := range utxos {
		if utxo.IsSlp() {
			continue
		}
		totalValue += utxo.Input.Value
		numInputs++
	}
	return GetMaxSendFromCount(totalValue, numInputs)
}

func GetMaxSendFromCount(totalValue int64, numInputs int) int64 {
	maxSend := totalValue - FeeP2pkh1In1OutTx - int64(numInputs-1)*InputFeeP2PKH
	if maxSend < DustMinimumOutput {
		maxSend = 0
	}
	return maxSend
}

func GetMaxSendLikeTip(totalValue int64, numInputs int) int64 {
	maxSend := totalValue - FeeP2pkh1In1OutTx - int64(numInputs-1)*InputFeeP2PKH - FeeOpReturnLike
	if maxSend < DustMinimumOutput {
		maxSend = 0
	}
	return maxSend
}

func GetMaxSendVoteTip(totalValue int64, numInputs int) int64 {
	maxSend := totalValue - FeeP2pkh1In1OutTx - int64(numInputs-1)*InputFeeP2PKH - FeeOpReturnVote
	if maxSend < DustMinimumOutput {
		maxSend = 0
	}
	return maxSend
}

func GetMaxSpendBuyTokens(totalValue int64, numInputs int) int64 {
	maxSpend := totalValue + DustMinimumOutput - FeeP2pkh1In2OutTx - int64(numInputs)*InputFeeP2PKH - FeeOpReturnSlpSend
	if maxSpend < DustMinimumOutput {
		maxSpend = 0
	}
	return maxSpend
}

func GetMaxSpendTopicMessage(totalValue int64, numInputs int) int64 {
	maxSpend := totalValue - FeeP2pkh1In1OutTx - int64(numInputs-1)*InputFeeP2PKH - FeeOpReturnTopic
	if maxSpend < DustMinimumOutput {
		maxSpend = 0
	}
	return maxSpend
}

func GetMaxSpendReply(totalValue int64, numInputs int) int64 {
	maxSpend := totalValue - FeeP2pkh1In1OutTx - int64(numInputs-1)*InputFeeP2PKH - FeeOpReturnReply
	if maxSpend < DustMinimumOutput {
		maxSpend = 0
	}
	return maxSpend
}
