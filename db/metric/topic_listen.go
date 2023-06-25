package metric

type TopicListen struct {
	Topic    string
}

func (s TopicListen) GetFields() map[string]interface{} {
	return map[string]interface{}{
		FieldQuantity: 1,
	}
}

func (s TopicListen) GetTags() map[string]string {
	return map[string]string{
		TagTopic: s.Topic,
	}
}

func AddTopicListen(request TopicListen) {
	writer := getInfluxWriter()
	if writer == nil {
		return
	}
	writer.Write(Point{
		Measurement: NameTopicListen,
		Fields:      request.GetFields(),
		Tags:        request.GetTags(),
	})
}
