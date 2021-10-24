package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type PollVote struct {
	PollOptionTxHash []byte
	Message          string
}

func (v PollVote) Get() ([]byte, error) {
	message := []byte(v.Message)
	if len(v.PollOptionTxHash) != memo.TxHashLength {
		return nil, jerr.Newf("invalid poll option tx hash length (%d)", len(v.PollOptionTxHash))
	}
	if len(message) > memo.MaxVoteCommentSize {
		return nil, jerr.New("message data too large")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixPollVote).
		AddData(v.PollOptionTxHash).
		AddData(message).
		Script()
	if err != nil {
		return nil, jerr.Get("error building poll vote script", err)
	}
	return pkScript, nil
}

func (v PollVote) Type() memo.OutputType {
	return memo.OutputTypeMemoPollVote
}
