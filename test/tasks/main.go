package tasks

import "github.com/memocash/index/test/suite"

const (
	TestSaveMessage    = "save_message"
	TestQueue          = "queue"
	TestQueueWait      = "queue_wait"
	TestSlpValidity    = "slp_validity"
	TestOpReturnSender = "op_return_sender"
	TestP2sh32         = "p2sh32"
)

func GetTests() []suite.Test {
	return []suite.Test{
		SaveMessage,
		queueTest,
		SlpValidity,
		OpReturnSender,
		P2sh32,
	}
}
