package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type TopicFollow struct {
	TopicName string
	Unfollow  bool
}

func (f TopicFollow) Get() ([]byte, error) {
	topicName := []byte(f.TopicName)
	if len(topicName) > memo.MaxTagMessageSize {
		return nil, fmt.Errorf("topic name too large")
	}
	if len(topicName) == 0 {
		return nil, fmt.Errorf("empty topic name")
	}
	var prefix []byte
	if f.Unfollow {
		prefix = memo.PrefixTopicUnfollow
	} else {
		prefix = memo.PrefixTopicFollow
	}
	pkScript, err := memo.GetBaseOpReturn().
		AddData(prefix).
		AddData(topicName).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building topic follow script; %w", err)
	}
	return pkScript, nil
}

func (f TopicFollow) Type() memo.OutputType {
	if f.Unfollow {
		return memo.OutputTypeMemoTopicUnfollow
	}
	return memo.OutputTypeMemoTopicFollow
}
