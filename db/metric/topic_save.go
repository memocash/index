package metric

type TopicSave struct {
	Topic    string
	Quantity int
}

func (s TopicSave) GetFields() map[string]interface{} {
	return map[string]interface{}{
		FieldQuantity: s.Quantity,
	}
}

func (s TopicSave) GetTags() map[string]string {
	return map[string]string{
		TagTopic: s.Topic,
	}
}

func AddTopicSave(request TopicSave) {
	writer := getInfluxWriter()
	if writer == nil {
		return
	}
	writer.Write(Point{
		Measurement: NameTopicSave,
		Fields:      request.GetFields(),
		Tags:        request.GetTags(),
	})
}
