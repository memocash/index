package metric

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"time"
)

type CollectionTopicSave struct {
	TopicSaves []*TopicSave
}

func (c *CollectionTopicSave) Add(topic string) {
	var found bool
	for _, topicSave := range c.TopicSaves {
		if topicSave.Topic == topic {
			topicSave.Quantity++
			found = true
			break
		}
	}
	if !found {
		c.TopicSaves = append(c.TopicSaves, &TopicSave{
			Topic:    topic,
			Quantity: 1,
		})
	}
}

func (c *CollectionTopicSave) Save() error {
	for _, topicSave := range c.TopicSaves {
		if err := AddTopicSave(*topicSave); err != nil {
			return fmt.Errorf("error adding topic save metric; %w", err)
		}
	}
	return nil
}

type TopicSave struct {
	Topic    string
	Quantity int
}

func (s TopicSave) GetFields() map[string]interface{} {
	return map[string]interface{}{
		TagQuantity: s.Quantity,
	}
}

func (s TopicSave) GetTags() map[string]string {
	return map[string]string{
		TagTopic: s.Topic,
	}
}

func AddTopicSave(request TopicSave) error {
	writeAPI, err := getInflux()
	if err != nil {
		return fmt.Errorf("error getting influx; %w", err)
	}
	if writeAPI == nil {
		return nil
	}
	fields := request.GetFields()
	tags := request.GetTags()
	p := influxdb2.NewPoint(NameTopicSave, tags, fields, time.Now())
	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		return fmt.Errorf("cannot write point; %w", err)
	}
	return nil
}
