package memo

const (
	CodePrefix = 0x6d

	CodeTest              = 0x00
	CodeSetName           = 0x01
	CodePost              = 0x02
	CodeReply             = 0x03
	CodeLike              = 0x04
	CodeSetProfile        = 0x05
	CodeFollow            = 0x06
	CodeUnfollow          = 0x07
	CodeSetImageBaseUrl   = 0x08
	CodeAttachPicture     = 0x09
	CodeSetProfilePicture = 0x0a
	CodeRepost            = 0x0b
	CodeTopicMessage      = 0x0c
	CodeTopicFollow       = 0x0d
	CodeTopicUnfollow     = 0x0e

	CodePollCreate = 0x10
	CodePollOption = 0x13
	CodePollVote   = 0x14

	CodeMute   = 0x16
	CodeUnMute = 0x17

	CodeLinkRequest = 0x20
	CodeLinkAccept  = 0x21
	CodeLinkRevoke  = 0x22

	CodeSendMoney = 0x24

	CodeSetAlias = 0x26

	CodeSellTokenMake      = 0x30
	CodeSellTokenOffer     = 0x31
	CodeSellTokenSignature = 0x32

	CodeTokenPin = 0x35
)

var (
	PrefixSetName         = []byte{CodePrefix, CodeSetName}
	PrefixPost            = []byte{CodePrefix, CodePost}
	PrefixReply           = []byte{CodePrefix, CodeReply}
	PrefixLike            = []byte{CodePrefix, CodeLike}
	PrefixSetProfile      = []byte{CodePrefix, CodeSetProfile}
	PrefixFollow          = []byte{CodePrefix, CodeFollow}
	PrefixUnfollow        = []byte{CodePrefix, CodeUnfollow}
	PrefixSetImageBaseUrl = []byte{CodePrefix, CodeSetImageBaseUrl}
	PrefixAttachPicture   = []byte{CodePrefix, CodeAttachPicture}
	PrefixSetProfilePic   = []byte{CodePrefix, CodeSetProfilePicture}
	PrefixRepost          = []byte{CodePrefix, CodeRepost}
	PrefixTopicMessage    = []byte{CodePrefix, CodeTopicMessage}
	PrefixTopicFollow     = []byte{CodePrefix, CodeTopicFollow}
	PrefixTopicUnfollow   = []byte{CodePrefix, CodeTopicUnfollow}
	PrefixPollCreate      = []byte{CodePrefix, CodePollCreate}
	PrefixPollOption      = []byte{CodePrefix, CodePollOption}
	PrefixPollVote        = []byte{CodePrefix, CodePollVote}
	PrefixSendMoney       = []byte{CodePrefix, CodeSendMoney}
	PrefixMute            = []byte{CodePrefix, CodeMute}
	PrefixUnmute          = []byte{CodePrefix, CodeUnMute}

	PrefixLinkRequest = []byte{CodePrefix, CodeLinkRequest}
	PrefixLinkAccept  = []byte{CodePrefix, CodeLinkAccept}
	PrefixLinkRevoke  = []byte{CodePrefix, CodeLinkRevoke}
	PrefixSetAlias    = []byte{CodePrefix, CodeSetAlias}

	PrefixBitcom = []byte("19HxigV4QyBv3tHpQVcUEQyq1pzZVdoAut")

	PrefixSlp = append([]byte("SLP"), 0x00)

	PrefixSellTokenMake      = []byte{CodePrefix, CodeSellTokenMake}
	PrefixSellTokenOffer     = []byte{CodePrefix, CodeSellTokenOffer}
	PrefixSellTokenSignature = []byte{CodePrefix, CodeSellTokenSignature}
	PrefixTokenPin           = []byte{CodePrefix, CodeTokenPin}
)

const (
	CodePollTypeSingle = 0x01
	CodePollTypeMulti  = 0x02
	CodePollTypeRank   = 0x03
)

func GetAllCodes() [][]byte {
	return append(GetMemoCodes(), [][]byte{
		PrefixBitcom,
		PrefixSlp,
		PrefixSellTokenMake,
		PrefixSellTokenOffer,
		PrefixSellTokenSignature,
	}...)
}

func GetMemoCodes() [][]byte {
	return [][]byte{
		PrefixSetName,
		PrefixPost,
		PrefixReply,
		PrefixLike,
		PrefixSetProfile,
		PrefixFollow,
		PrefixUnfollow,
		PrefixSetImageBaseUrl,
		PrefixAttachPicture,
		PrefixSetProfilePic,
		PrefixRepost,
		PrefixTopicMessage,
		PrefixPollCreate,
		PrefixPollOption,
		PrefixPollVote,
		PrefixTopicFollow,
		PrefixTopicUnfollow,
		PrefixSendMoney,
		PrefixMute,
		PrefixUnmute,
		PrefixTokenPin,
		PrefixLinkRequest,
		PrefixLinkAccept,
		PrefixLinkRevoke,
		PrefixSetAlias,
	}
}

func IsMemo(prefix []byte) bool {
	return len(prefix) == 2 && prefix[0] == CodePrefix
}
