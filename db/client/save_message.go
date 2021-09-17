package client

type SaveMessage struct {
}

func (m *SaveMessage) Save() error {
	return nil
}

func NewSaveMessage() *SaveMessage {
	return &SaveMessage{}
}
