package topic

import (
	"encoding/hex"
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

var itemRoute = admin.Route{
	Pattern: admin.UrlTopicItem,
	Handler: func(r admin.Response) {
		var topicItemRequest = new(admin.TopicItemRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(topicItemRequest); err != nil {
			jerr.Get("error unmarshalling topic item request", err).Print()
			return
		}
		uid, err := hex.DecodeString(topicItemRequest.Uid)
		if err != nil {
			jerr.Get("error parsing uid for topic item", err)
		}
		var topicItemResponse = new(admin.TopicItemResponse)
		shardConfig := config.GetShardConfig(uint32(topicItemRequest.Shard), config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetSingle(topicItemRequest.Topic, uid); err != nil {
			jerr.Get("error getting topic item for admin view", err).Print()
			return
		}
		if len(db.Messages) != 1 {
			jerr.Newf("error unexpected message count: %d", len(db.Messages))
			return
		}
		topicItemResponse.Item = admin.TopicItem{
			Topic:   topicItemRequest.Topic,
			Shard:   topicItemRequest.Shard,
			Uid:     hex.EncodeToString(db.Messages[0].Uid),
			Message: hex.EncodeToString(db.Messages[0].Message),
		}
		if err := json.NewEncoder(r.Writer).Encode(topicItemResponse); err != nil {
			jerr.Get("error writing json topic item response data", err).Print()
			return
		}
	},
}
