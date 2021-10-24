package memo

const (
	MaxFundAmount = 10000
	MinFundAmount = 1250

	MaxFaucetTransactionsDay  = 50
	MaxFaucetTransactionsHour = 15

	DesiredFaucetUtxoCount = 15
)

func IsFreeTx(outputs []*Output) bool {
	if len(outputs) != 1 {
		return false
	}
	switch outputs[0].GetType() {
	case OutputTypeMemoMessage,
		OutputTypeMemoSetProfile,
		OutputTypeMemoSetName,
		OutputTypeMemoLike,
		OutputTypeMemoSetProfilePic,
		OutputTypeMemoTopicMessage,
		OutputTypeMemoReply,
		OutputTypeMemoFollow,
		OutputTypeMemoTopicFollow,
		OutputTypeLinkRequest,
		OutputTypeLinkAccept,
		OutputTypeMemoPollVote:
		return true
	case OutputTypeMemoUnfollow,
		OutputTypeMemoTopicUnfollow:
		return false
	}
	return false
}

func FreeTxFee(outputs []*Output) int64 {
	if len(outputs) != 1 {
		return 0
	}
	outputValuePlusFee, err := outputs[0].GetValuePlusFee()
	if err != nil {
		return 0
	}
	return BaseTxFee + InputFeeP2PKH + OutputFeeP2PKH + outputValuePlusFee
}
