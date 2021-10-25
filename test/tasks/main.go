package tasks

import "github.com/memocash/server/test/suite"

const (
	TestDoubleSpend = "double_spend"
	TestSaveMessage = "save_message"
)

func GetTests() []suite.Test {
	return []suite.Test{
		SaveMessage,
		doubleSpendTest,
	}
}
