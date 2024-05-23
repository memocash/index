package script

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type TopicMessage struct {
	TopicName string
	Message   string
}

func (p TopicMessage) Get() ([]byte, error) {
	topicName := []byte(p.TopicName)
	message := []byte(p.Message)
	if len(topicName)+len(message) > memo.MaxTagMessageSize {
		return nil, fmt.Errorf("data too large")
	}
	if len(topicName) == 0 || len(message) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixTopicMessage).
		AddData(topicName).
		AddData(message).
		Script()
	if err != nil {
		return nil, fmt.Errorf("error building topic message script; %w", err)
	}
	return pkScript, nil
}

func (p TopicMessage) Type() memo.OutputType {
	return memo.OutputTypeMemoTopicMessage
}
