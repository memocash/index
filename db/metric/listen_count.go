package metric

type ListenCount struct {
	Quantity int
}

func (c ListenCount) GetFields() map[string]interface{} {
	return map[string]interface{}{
		FieldQuantity: c.Quantity,
	}
}

func AddListenCount(request ListenCount) {
	writer := getInfluxWriter()
	if writer == nil {
		return
	}
	writer.Write(Point{
		Measurement: NameListenCount,
		Fields:      request.GetFields(),
	})
}
