package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type PollCreate struct {
	Question    string
	OptionCount int
	PollType    memo.PollType
}

func (c PollCreate) Get() ([]byte, error) {
	question := []byte(c.Question)
	if len(question) > memo.MaxPollQuestionSize {
		return nil, fmt.Errorf("error poll question too big (%d)", len(question))
	}
	if len(question) == 0 {
		return nil, fmt.Errorf("empty question")
	}
	if c.OptionCount == 0 {
		return nil, fmt.Errorf("empty option count")
	}
	var pollType byte
	switch c.PollType {
	case memo.PollTypeOne:
		pollType = memo.CodePollTypeSingle
	case memo.PollTypeAny:
		pollType = memo.CodePollTypeMulti
	default:
		return nil, fmt.Errorf("invalid poll type (%s)", c.PollType)
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixPollCreate).
		AddData([]byte{pollType}).
		AddData([]byte{byte(c.OptionCount)}).
		AddData(question).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building poll create script; %w", err)
	}
	return pkScript, nil
}

func (c PollCreate) Type() memo.OutputType {
	switch c.PollType {
	case memo.PollTypeOne:
		return memo.OutputTypeMemoPollQuestionSingle
	case memo.PollTypeAny:
		return memo.OutputTypeMemoPollQuestionMulti
	default:
		return memo.OutputTypeUnknown
	}
}
