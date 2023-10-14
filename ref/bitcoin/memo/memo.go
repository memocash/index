package memo

const (
	MaxPostSize         = 65000
	MaxReplySize        = 65000
	MaxTagMessageSize   = 65000
	MaxPollQuestionSize = 209
	MaxPollOptionSize   = 184
	MaxVoteCommentSize  = 184
	MaxSendSize         = 65000
	MaxFileSize         = 99000

	OldMaxPostSize       = 217
	OldMaxReplySize      = 184
	OldMaxTagMessageSize = 204
	OldMaxSendSize       = 194
)

const (
	PkHashLength     = 20
	ScriptHashLength = 20
	TxHashLength     = 32
	PubKeyLength     = 64
	LockHashLength   = TxHashLength
	AddressLength    = 25

	AddressStringLength = 35
	TxStringLength      = 64

	BlockHeaderLength = 80
)

// https://bitcoin.stackexchange.com/questions/1195/how-to-calculate-transaction-size-before-sending-legacy-non-segwit-p2pkh-p2sh
const (
	MaxTxFee          int64 = 425
	OutputFeeP2PKH    int64 = 34
	OutputFeeP2SH     int64 = 32
	OutputBaseFee     int64 = 12
	OutputFeeOpReturn int64 = 20
	OutputOpDataFee   int64 = 3
	OutputValueSize   int64 = 8
	InputFeeP2PKH     int64 = 148
	BaseTxFee         int64 = 10

	FeeOpReturnSlpSend int64 = 72
	FeeOpReturnLike    int64 = 55
	FeeOpReturnVote    int64 = 55
	FeeOpReturnTopic   int64 = 23
	FeeOpReturnReply   int64 = 55

	FeeP2pkh1In1OutTx   = BaseTxFee + InputFeeP2PKH + OutputFeeP2PKH
	FeeP2pkh1In2OutTx   = BaseTxFee + InputFeeP2PKH + OutputFeeP2PKH*2
	Fee2In3OutSlpSendTx = BaseTxFee + InputFeeP2PKH*2 + OutputFeeP2PKH*3 + FeeOpReturnSlpSend
)

const (
	Int2Size = 2 // int16
	Int4Size = 4 // int32
	Int8Size = 8 // int64
)

const MaxAncestors = 2000
const FaucetMaxAncestors = 20

const DustMinimumOutput int64 = 546

const BeginningOfMemoHeight = 525000
const BeginningOfSlpHeight = 543000
const BeginningOfSellToken = 585000
const BsvForkHeight = 555766

type PollType string

const (
	PollTypeOne  PollType = "one"
	PollTypeAny  PollType = "any"
	PollTypeRank PollType = "rank"
)

func IsSpendable(balance int64) bool {
	return balance > DustMinimumOutput+FeeP2pkh1In1OutTx
}

func IsSpendableTokenPin(balance int64) bool {
	return balance > DustMinimumOutput*2+Fee2In3OutSlpSendTx
}
