package topic

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/admin/admin"
)

var listRoute = admin.Route{
	Pattern: admin.UrlTopicList,
	Handler: func(r admin.Response) {
		jlog.Logf("Here")
		var topicListResponse = &admin.TopicListResponse{
			Topics: []admin.Topic{{
				Name: "example",
			}, {
				Name: "example2",
			}},
		}
		if err := json.NewEncoder(r.Writer).Encode(topicListResponse); err != nil {
			jerr.Get("error writing json topic list response data", err).Print()
			return
		}
	},
}
