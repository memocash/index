package tasks

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/test/suite"
	"time"
)

var SaveMessage = suite.Test{
	Name: "save_message",
	Test: func(request *suite.TestRequest) error {
		var messages = []item.Object{&item.Message{
			Id:      0,
			Message: "Test!",
			Created: time.Now(),
		}}
		if err := item.Save(messages); err != nil {
			return jerr.Get("error saving message to client", err)
		}
		return nil
	},
}
