package memo

import (
	"bytes"
	"github.com/jchavannes/btcd/txscript"
)

type OutputType uint

func (s OutputType) String() string {
	var outputString, ok = outputTypeStringMap[s]
	if !ok {
		return StringUnknown
	}
	return outputString
}

const (
	OutputTypeP2PKH OutputType = iota
	OutputTypeP2SH
	OutputTypeP2PK
	OutputTypeReturn
	OutputTypeMemoMessage
	OutputTypeMemoSetName
	OutputTypeMemoFollow
	OutputTypeMemoUnfollow
	OutputTypeMemoLike
	OutputTypeMemoReply
	OutputTypeMemoSetProfile
	OutputTypeMemoTopicMessage
	OutputTypeMemoTopicFollow
	OutputTypeMemoTopicUnfollow
	OutputTypeMemoPollQuestionSingle
	OutputTypeMemoPollQuestionMulti
	OutputTypeMemoPollOption
	OutputTypeMemoPollVote
	OutputTypeMemoSetProfilePic
	OutputTypeMemoSend
	OutputTypeBitcom
	OutputTypeTokenCreate
	OutputTypeTokenCreateNftGroup
	OutputTypeTokenCreateNftChild
	OutputTypeTokenMint
	OutputTypeTokenMintNftGroup
	OutputTypeTokenSend
	OutputTypeTokenSendNftGroup
	OutputTypeTokenSendNftChild
	OutputTypeTokenSell
	OutputTypeTokenOffer
	OutputTypeTokenSignature
	OutputTypeMemoMute
	OutputTypeMemoUnMute
	OutputTypeTokenPin
	OutputTypeLinkRequest
	OutputTypeLinkAccept
	OutputTypeLinkRevoke
	OutputTypeSetAlias
	OutputTypeUnknown
	OutputTypeNone
)

const (
	StringP2pkh             = "p2pkh"
	StringP2sh              = "p2sh"
	StringP2pk              = "p2pk"
	StringReturn            = "return"
	StringMemoMessage       = "memo-message"
	StringMemoSetName       = "memo-set-name"
	StringMemoFollow        = "memo-follow"
	StringMemoUnfollow      = "memo-unfollow"
	StringMemoLike          = "memo-like"
	StringMemoReply         = "memo-reply"
	StringMemoSetProfile    = "memo-set-profile"
	StringMemoSetProfilePic = "memo-set-profile-pic"
	StringMemoSend          = "memo-send"
	StringMemoMute          = "memo-mute"
	StringMemoUnmute        = "memo-unmute"
	StringBitcom            = "bitcom"
	StringMemoTopicMessage  = "topic-message"
	StringMemoTopicFollow   = "topic-follow"
	StringMemoTopicUnfollow = "topic-unfollow"
	StringMemoPollQuestion  = "poll-question"
	StringMemoPollOption    = "poll-option"
	StringMemoPollVote      = "poll-vote"
	StringTokenCreate       = "token-create"
	StringTokenMint         = "token-mint"
	StringTokenSend         = "token-send"
	StringTokenSell         = "token-sell"
	StringTokenOffer        = "token-offer"
	StringTokenSignature    = "token-signature"
	StringTokenPin          = "token-pin"
	StringLinkRequest       = "link-request"
	StringLinkAccept        = "link-accept"
	StringLinkRevoke        = "link-revoke"
	StringSetAlias          = "set-alias"
	StringUnknown           = "unknown"
	StringNone              = "-"
)

var outputTypeStringMap = map[OutputType]string{
	OutputTypeP2PKH:                  StringP2pkh,
	OutputTypeP2SH:                   StringP2sh,
	OutputTypeP2PK:                   StringP2pk,
	OutputTypeReturn:                 StringReturn,
	OutputTypeMemoMessage:            StringMemoMessage,
	OutputTypeMemoSetName:            StringMemoSetName,
	OutputTypeMemoFollow:             StringMemoFollow,
	OutputTypeMemoUnfollow:           StringMemoUnfollow,
	OutputTypeMemoLike:               StringMemoLike,
	OutputTypeMemoReply:              StringMemoReply,
	OutputTypeMemoSetProfile:         StringMemoSetProfile,
	OutputTypeMemoTopicMessage:       StringMemoTopicMessage,
	OutputTypeMemoTopicFollow:        StringMemoTopicFollow,
	OutputTypeMemoTopicUnfollow:      StringMemoTopicUnfollow,
	OutputTypeMemoPollQuestionSingle: StringMemoPollQuestion,
	OutputTypeMemoPollQuestionMulti:  StringMemoPollQuestion,
	OutputTypeMemoPollOption:         StringMemoPollOption,
	OutputTypeMemoPollVote:           StringMemoPollVote,
	OutputTypeMemoSetProfilePic:      StringMemoSetProfilePic,
	OutputTypeMemoSend:               StringMemoSend,
	OutputTypeMemoMute:               StringMemoMute,
	OutputTypeMemoUnMute:             StringMemoUnmute,
	OutputTypeBitcom:                 StringBitcom,

	OutputTypeTokenCreate:         StringTokenCreate,
	OutputTypeTokenCreateNftGroup: StringTokenCreate,
	OutputTypeTokenCreateNftChild: StringTokenCreate,
	OutputTypeTokenMint:           StringTokenMint,
	OutputTypeTokenMintNftGroup:   StringTokenMint,
	OutputTypeTokenSend:           StringTokenSend,
	OutputTypeTokenSendNftGroup:   StringTokenSend,
	OutputTypeTokenSendNftChild:   StringTokenSend,
	OutputTypeNone:                StringNone,

	OutputTypeLinkRequest: StringLinkRequest,
	OutputTypeLinkAccept:  StringLinkAccept,
	OutputTypeLinkRevoke:  StringLinkRevoke,
	OutputTypeSetAlias:    StringSetAlias,

	OutputTypeTokenSell:      StringTokenSell,
	OutputTypeTokenOffer:     StringTokenOffer,
	OutputTypeTokenSignature: StringTokenSignature,
	OutputTypeTokenPin:       StringTokenPin,
}

var outputTypePrefixMap = map[OutputType][][]byte{
	OutputTypeMemoMessage:       {PrefixPost},
	OutputTypeMemoReply:         {PrefixReply},
	OutputTypeMemoLike:          {PrefixLike},
	OutputTypeMemoSetName:       {PrefixSetName},
	OutputTypeMemoSetProfile:    {PrefixSetProfile},
	OutputTypeMemoSetProfilePic: {PrefixSetProfilePic},
	OutputTypeMemoFollow:        {PrefixFollow},
	OutputTypeMemoUnfollow:      {PrefixUnfollow},
	OutputTypeMemoMute:          {PrefixMute},
	OutputTypeMemoUnMute:        {PrefixUnmute},

	OutputTypeMemoTopicMessage:  {PrefixTopicMessage},
	OutputTypeMemoTopicFollow:   {PrefixTopicFollow},
	OutputTypeMemoTopicUnfollow: {PrefixTopicUnfollow},

	OutputTypeMemoPollOption: {PrefixPollOption},
	OutputTypeMemoPollVote:   {PrefixPollVote},
	OutputTypeMemoSend:       {PrefixSendMoney},
	OutputTypeBitcom:         {PrefixBitcom},

	OutputTypeLinkRequest: {PrefixLinkRequest},
	OutputTypeLinkAccept:  {PrefixLinkAccept},
	OutputTypeLinkRevoke:  {PrefixLinkRevoke},
	OutputTypeSetAlias:    {PrefixSetAlias},

	OutputTypeMemoPollQuestionSingle: {PrefixPollCreate, []byte{CodePollTypeSingle}},
	OutputTypeMemoPollQuestionMulti:  {PrefixPollCreate, []byte{CodePollTypeMulti}},

	OutputTypeTokenCreate:         {PrefixSlp, []byte{SlpDefaultTokenType}, []byte(SlpTxTypeGenesis)},
	OutputTypeTokenCreateNftGroup: {PrefixSlp, []byte{SlpNftGroupTokenType}, []byte(SlpTxTypeGenesis)},
	OutputTypeTokenCreateNftChild: {PrefixSlp, []byte{SlpNftChildTokenType}, []byte(SlpTxTypeGenesis)},
	OutputTypeTokenMint:           {PrefixSlp, []byte{SlpDefaultTokenType}, []byte(SlpTxTypeMint)},
	OutputTypeTokenMintNftGroup:   {PrefixSlp, []byte{SlpNftGroupTokenType}, []byte(SlpTxTypeMint)},
	OutputTypeTokenSend:           {PrefixSlp, []byte{SlpDefaultTokenType}, []byte(SlpTxTypeSend)},
	OutputTypeTokenSendNftGroup:   {PrefixSlp, []byte{SlpNftGroupTokenType}, []byte(SlpTxTypeSend)},
	OutputTypeTokenSendNftChild:   {PrefixSlp, []byte{SlpNftChildTokenType}, []byte(SlpTxTypeSend)},

	OutputTypeTokenSell:      {PrefixSellTokenMake},
	OutputTypeTokenOffer:     {PrefixSellTokenOffer},
	OutputTypeTokenSignature: {PrefixSellTokenSignature},
	OutputTypeTokenPin:       {PrefixTokenPin},
}

func IsOfOutputType(item OutputType, acceptable []OutputType) bool {
	for _, accept := range acceptable {
		if item == accept {
			return true
		}
	}
	return false
}

func GetOutputTypeNew(pkScript []byte) OutputType {
	if len(pkScript) < 3 || pkScript[0] != txscript.OP_RETURN {
		return OutputTypeUnknown
	}
	potentialPrefix := pkScript[2:]
loop:
	for outputType, multiPrefix := range outputTypePrefixMap {
		var tmpPotentialPrefix = potentialPrefix
		for _, push := range multiPrefix {
			if len(tmpPotentialPrefix) <= len(push) || !bytes.Equal(tmpPotentialPrefix[:len(push)], push) {
				continue loop
			}
			if len(tmpPotentialPrefix) >= len(push)+1 {
				tmpPotentialPrefix = tmpPotentialPrefix[len(push)+1:]
			}
		}
		return outputType
	}
	return OutputTypeUnknown
}

func GetOutputType(pkScript []byte) OutputType {
	pushData, err := txscript.PushedData(pkScript)
	var lenPushData = len(pushData)
	if err != nil || lenPushData < 1 {
		return OutputTypeUnknown
	}
	if len(pushData[0]) == 0 && len(pushData) > 1 {
		pushData = pushData[1:]
	}
loop:
	for outputType, multiPrefix := range outputTypePrefixMap {
		for i, push := range multiPrefix {
			if lenPushData <= i || !bytes.Equal(pushData[i], push) {
				continue loop
			}
		}
		return outputType
	}
	return OutputTypeUnknown
}
