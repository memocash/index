package tasks

import "github.com/memocash/index/test/suite"

const (
	TestSaveMessage = "save_message"
	TestQueue       = "queue"
	TestQueueWait   = "queue_wait"
)

func GetTests() []suite.Test {
	return []suite.Test{
		SaveMessage,
		queueTest,
		waitTest,
	}
}
