package topic

import (
	"encoding/hex"
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
	"reflect"
	"strconv"
)

var typeOfBytes = reflect.TypeOf([]byte(nil))

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
		var props = make(map[string]interface{})
		for _, topic := range item.GetTopics() {
			if topic.GetTopic() != topicItemRequest.Topic {
				continue
			}
			objType := reflect.ValueOf(topic).Elem().Type()
			obj := reflect.New(objType).Interface().(item.Object)
			obj.SetUid(db.Messages[0].Uid)
			obj.Deserialize(db.Messages[0].Message)
			objVal := reflect.ValueOf(obj)
			for i := 0; i < objType.NumField(); i++ {
				fieldValue := objVal.Elem().Field(i)
				fieldName := objVal.Elem().Type().Field(i).Name
				switch fieldValue.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					props[fieldName] = strconv.FormatInt(fieldValue.Int(), 10)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					props[fieldName] = strconv.FormatUint(fieldValue.Uint(), 10)
				case reflect.Slice:
					if fieldValue.Type() == typeOfBytes {
						/*var buf = new(bytes.Buffer)
						if err := binary.Read(buf, binary.BigEndian, fieldValue.Addr()); err != nil {
							jerr.Get("error reading byte slice buffer for topic item", err).Print()
							return
						}
						props[fieldName], _ = ioutil.ReadAll(buf)*/
					}
					props[fieldName] = fieldValue.String()
				case reflect.String:
					props[fieldName] = fieldValue.String()
				default:
					props[fieldName] = fieldValue.String()
				}
			}
		}
		topicItemResponse.Item = admin.TopicItem{
			Topic:   topicItemRequest.Topic,
			Shard:   topicItemRequest.Shard,
			Uid:     hex.EncodeToString(db.Messages[0].Uid),
			Message: hex.EncodeToString(db.Messages[0].Message),
			Props:   props,
		}
		if err := json.NewEncoder(r.Writer).Encode(topicItemResponse); err != nil {
			jerr.Get("error writing json topic item response data", err).Print()
			return
		}
	},
}
