package tasks

import "github.com/memocash/server/test/suite"

func GetTests() []suite.Test {
	return []suite.Test{
		SaveMessage,
	}
}
