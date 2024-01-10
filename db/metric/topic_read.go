package metric

type TopicRead struct {
	Topic    string
	Quantity int
}

func (s TopicRead) GetFields() map[string]interface{} {
	return map[string]interface{}{
		FieldQuantity: s.Quantity,
	}
}

func (s TopicRead) GetTags() map[string]string {
	return map[string]string{
		TagTopic: s.Topic,
	}
}

func AddTopicRead(request TopicRead) {
	writer := getInfluxWriter()
	if writer == nil {
		return
	}
	writer.Write(Point{
		Measurement: NameTopicRead,
		Fields:      request.GetFields(),
		Tags:        request.GetTags(),
	})
}
