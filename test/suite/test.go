package suite

type TestRequest struct {
	Suite *Suite
	Args  []string
}

type Test struct {
	Name string
	Test func(*TestRequest) error
}

func Tests(routeSets ...[]Test) []Test {
	var allTests []Test
	for _, routeSet := range routeSets {
		allTests = append(allTests, routeSet...)
	}
	return allTests
}
