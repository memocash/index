package script

import (
	"github.com/jchavannes/jgo/jerr"
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
		return nil, jerr.New("data too large")
	}
	if len(topicName) == 0 || len(message) == 0 {
		return nil, jerr.New("empty data")
	}

	pkScript, err := memo.GetBaseOpReturn().
		AddData(memo.PrefixTopicMessage).
		AddData(topicName).
		AddData(message).
		Script()
	if err != nil {
		return nil, jerr.Get("error building topic message script", err)
	}
	return pkScript, nil
}

func (p TopicMessage) Type() memo.OutputType {
	return memo.OutputTypeMemoTopicMessage
}
