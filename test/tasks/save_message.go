package tasks

import (
	"fmt"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/test/suite"
	"time"
)

const TestMessage = "Test!"

var SaveMessage = suite.Test{
	Name: TestSaveMessage,
	Test: func(request *suite.TestRequest) error {
		var messages = []db.Object{&item.Message{
			Id:      0,
			Message: TestMessage,
			Created: time.Now(),
		}}
		if err := db.Save(messages); err != nil {
			return fmt.Errorf("error saving message to client; %w", err)
		}
		message, err := item.GetMessage(0)
		if err != nil {
			return fmt.Errorf("error getting message from client; %w", err)
		}
		if message.Message != TestMessage {
			return fmt.Errorf("error message unexpected: %s", message.Message)
		}
		return nil
	},
}
