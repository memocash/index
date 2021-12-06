package tasks

import "github.com/memocash/index/test/suite"

const (
	TestDoubleSpend = "double_spend"
	TestSaveMessage = "save_message"
	TestQueue       = "queue"
	TestQueueWait   = "queue_wait"
)

func GetTests() []suite.Test {
	return []suite.Test{
		SaveMessage,
		doubleSpendTest,
		queueTest,
		waitTest,
	}
}
