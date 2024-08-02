package topic

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"log"
)

var viewRoute = admin.Route{
	Pattern: admin.UrlTopicView,
	Handler: func(r admin.Response) {
		var topicViewRequest = new(admin.TopicViewRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(topicViewRequest); err != nil {
			log.Printf("error unmarshalling topic view request; %v", err)
			return
		}
		var start []byte
		if topicViewRequest.Start != "" {
			var err error
			if start, err = hex.DecodeString(topicViewRequest.Start); err != nil {
				fmt.Errorf("error parsing start from topic view request; %w", err)
				return
			}
		}
		var topicViewResponse = new(admin.TopicViewResponse)
		for _, shardConfig := range config.GetQueueShards() {
			if topicViewRequest.Shard >= 0 && uint32(topicViewRequest.Shard) != shardConfig.Shard {
				continue
			}
			db := client.NewClient(shardConfig.GetHost())
			if err := db.GetByPrefix(r.Request.Context(), topicViewRequest.Topic, client.NewStart(start)); err != nil {
				log.Printf("error getting topic items for admin view; %v", err)
				return
			}
			for _, msg := range db.Messages {
				topicViewResponse.Items = append(topicViewResponse.Items, admin.TopicItem{
					Topic:   topicViewRequest.Topic,
					Uid:     hex.EncodeToString(msg.Uid),
					Message: hex.EncodeToString(msg.Message),
					Shard:   uint(shardConfig.Shard),
				})
			}
		}
		if err := json.NewEncoder(r.Writer).Encode(topicViewResponse); err != nil {
			log.Printf("error writing json topic view response data; %v", err)
			return
		}
	},
}
