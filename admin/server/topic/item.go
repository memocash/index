package topic

import (
	"encoding/hex"
	"encoding/json"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var typeOfBytes = reflect.TypeOf([]byte(nil))
var typeOfBytes25 = reflect.TypeOf([25]byte{})
var typeOfBytes32 = reflect.TypeOf([32]byte{})

var itemRoute = admin.Route{
	Pattern: admin.UrlTopicItem,
	Handler: func(r admin.Response) {
		var topicItemRequest = new(admin.TopicItemRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(topicItemRequest); err != nil {
			log.Printf("error unmarshalling topic item request; %v", err)
			return
		}
		uid, err := hex.DecodeString(topicItemRequest.Uid)
		if err != nil {
			log.Printf("error parsing uid for topic item; %v", err)
		}
		var topicItemResponse = new(admin.TopicItemResponse)
		shardConfig := config.GetShardConfig(uint32(topicItemRequest.Shard), config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetSingle(r.Request.Context(), topicItemRequest.Topic, uid); err != nil {
			log.Printf("error getting topic item for admin view; %v", err)
			return
		}
		if len(dbClient.Messages) != 1 {
			log.Printf("error unexpected message count: %d", len(dbClient.Messages))
			return
		}
		var props = make(map[string]interface{})
		for _, topic := range item.GetTopicsSorted() {
			if topic.GetTopic() != topicItemRequest.Topic {
				continue
			}
			objType := reflect.ValueOf(topic).Elem().Type()
			obj := reflect.New(objType).Interface().(db.Object)
			obj.SetUid(dbClient.Messages[0].Uid)
			obj.Deserialize(dbClient.Messages[0].Message)
			elem := reflect.ValueOf(obj).Elem()
			for i := 0; i < objType.NumField(); i++ {
				fieldValue := elem.Field(i)
				fieldName := elem.Type().Field(i).Name
				switch fieldValue.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					props[fieldName] = strconv.FormatInt(fieldValue.Int(), 10)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					props[fieldName] = strconv.FormatUint(fieldValue.Uint(), 10)
				case reflect.Slice:
					if fieldValue.Type() == typeOfBytes {
						if strings.Contains(strings.ToLower(fieldName), "txhash") ||
							strings.Contains(strings.ToLower(fieldName), "blockhash") {
							props[fieldName] = hex.EncodeToString(jutil.ByteReverse(fieldValue.Bytes()))
						} else {
							props[fieldName] = hex.EncodeToString(fieldValue.Bytes())
						}
					} else {
						props[fieldName] = fieldValue.String()
					}
				case reflect.Array:
					if fieldValue.Type() == typeOfBytes25 {
						props[fieldName] = wallet.GetAddrFromBytes(fieldValue.Bytes()).String()
					} else if fieldValue.Type() == typeOfBytes32 {
						props[fieldName] = hex.EncodeToString(jutil.ByteReverse(fieldValue.Bytes()))
					} else {
						props[fieldName] = fieldValue.String()
					}
				case reflect.String:
					props[fieldName] = fieldValue.String()
				case reflect.Bool:
					props[fieldName] = fieldValue.Bool()
				default:
					switch v := fieldValue.Interface().(type) {
					case time.Time:
						props[fieldName] = v.Format(time.RFC3339Nano)
					default:
						props[fieldName] = fieldValue.String()
					}
				}
			}
		}
		topicItemResponse.Item = admin.TopicItem{
			Topic:   topicItemRequest.Topic,
			Shard:   topicItemRequest.Shard,
			Uid:     hex.EncodeToString(dbClient.Messages[0].Uid),
			Message: hex.EncodeToString(dbClient.Messages[0].Message),
			Props:   props,
		}
		if err := json.NewEncoder(r.Writer).Encode(topicItemResponse); err != nil {
			log.Printf("error writing json topic item response data; %v", err)
			return
		}
	},
}
