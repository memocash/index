package topic

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/item"
)

var listRoute = admin.Route{
	Pattern: admin.UrlTopicList,
	Handler: func(r admin.Response) {
		topics := item.GetTopicsSorted()
		var topicListResponse = new(admin.TopicListResponse)
		topicListResponse.Topics = make([]admin.Topic, len(topics))
		for i := range topics {
			topicListResponse.Topics[i] = admin.Topic{
				Name: topics[i].GetTopic(),
			}
		}
		if err := json.NewEncoder(r.Writer).Encode(topicListResponse); err != nil {
			jerr.Get("error writing json topic list response data", err).Print()
			return
		}
	},
}
