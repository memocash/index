package topic

import (
	"encoding/hex"
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

var viewRoute = admin.Route{
	Pattern: admin.UrlTopicView,
	Handler: func(r admin.Response) {
		var topicViewRequest = new(admin.TopicViewRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(topicViewRequest); err != nil {
			jerr.Get("error unmarshalling topic view request", err).Print()
			return
		}
		var start []byte
		if topicViewRequest.Start != "" {
			var err error
			if start, err = hex.DecodeString(topicViewRequest.Start); err != nil {
				jerr.Get("error parsing start from topic view request", err)
				return
			}
		}
		var topicViewResponse = new(admin.TopicViewResponse)
		for _, shardConfig := range config.GetQueueShards() {
			if topicViewRequest.Shard >= 0 && uint32(topicViewRequest.Shard) != shardConfig.Min {
				continue
			}
			db := client.NewClient(shardConfig.GetHost())
			if err := db.Get(topicViewRequest.Topic, start, false); err != nil {
				jerr.Get("error getting topic items for admin view", err).Print()
				return
			}
			for _, msg := range db.Messages {
				topicViewResponse.Items = append(topicViewResponse.Items, admin.TopicItem{
					Topic:   topicViewRequest.Topic,
					Uid:     hex.EncodeToString(msg.Uid),
					Message: hex.EncodeToString(msg.Message),
					Shard:   uint(shardConfig.Min),
				})
			}
		}
		if err := json.NewEncoder(r.Writer).Encode(topicViewResponse); err != nil {
			jerr.Get("error writing json topic view response data", err).Print()
			return
		}
	},
}
