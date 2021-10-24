package script

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type TopicFollow struct {
	TopicName string
	Unfollow  bool
}

func (f TopicFollow) Get() ([]byte, error) {
	topicName := []byte(f.TopicName)
	if len(topicName) > memo.MaxTagMessageSize {
		return nil, jerr.New("topic name too large")
	}
	if len(topicName) == 0 {
		return nil, jerr.New("empty topic name")
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
		return nil, jerr.Get("error building topic follow script", err)
	}
	return pkScript, nil
}

func (f TopicFollow) Type() memo.OutputType {
	if f.Unfollow {
		return memo.OutputTypeMemoTopicUnfollow
	}
	return memo.OutputTypeMemoTopicFollow
}
