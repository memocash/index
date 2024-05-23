package gen

import (
	"fmt"
)

var (
	NilInputGetterError      = fmt.Errorf("error nil input getter")
	NotEnoughValueError      = fmt.Errorf("error unable to find enough value to spend")
	NotEnoughTokenValueError = fmt.Errorf("error unable to find enough token value to spend")
	BelowDustLimitError      = fmt.Errorf("error output below dust limit")
)
