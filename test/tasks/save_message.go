package tasks

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/test/suite"
	"time"
)

const TestMessage = "Test!"

var SaveMessage = suite.Test{
	Name: "save_message",
	Test: func(request *suite.TestRequest) error {
		var messages = []item.Object{&item.Message{
			Id:      0,
			Message: TestMessage,
			Created: time.Now(),
		}}
		if err := item.Save(messages); err != nil {
			return jerr.Get("error saving message to client", err)
		}
		message, err := item.GetMessage(0)
		if err != nil {
			return jerr.Get("error getting message from client", err)
		}
		if message.Message != TestMessage {
			return jerr.Newf("error message unexpected: %s", message.Message)
		}
		return nil
	},
}
