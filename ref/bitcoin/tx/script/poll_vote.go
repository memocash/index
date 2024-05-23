package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type PollVote struct {
	PollOptionTxHash []byte
	Message          string
}

func (v PollVote) Get() ([]byte, error) {
	message := []byte(v.Message)
	if len(v.PollOptionTxHash) != memo.TxHashLength {
		return nil, fmt.Errorf("invalid poll option tx hash length (%d)", len(v.PollOptionTxHash))
	}
	if len(message) > memo.MaxVoteCommentSize {
		return nil, fmt.Errorf("message data too large")
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixPollVote).
		AddData(v.PollOptionTxHash).
		AddData(message).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building poll vote script; %w", err)
	}
	return pkScript, nil
}

func (v PollVote) Type() memo.OutputType {
	return memo.OutputTypeMemoPollVote
}
