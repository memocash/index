package gen

import (
	"github.com/jchavannes/jgo/jerr"
)

const (
	nilInputGetterErrorText      = "error nil input getter"
	NotEnoughValueErrorText      = "error unable to find enough value to spend"
	NotEnoughTokenValueErrorText = "error unable to find enough token value to spend"
	BelowDustLimitErrorText      = "error output below dust limit"
)

var (
	NilInputGetterError      = jerr.New(nilInputGetterErrorText)
	NotEnoughValueError      = jerr.New(NotEnoughValueErrorText)
	NotEnoughTokenValueError = jerr.New(NotEnoughTokenValueErrorText)
	BelowDustLimitError      = jerr.New(BelowDustLimitErrorText)
)

func IsNotEnoughValueError(err error) bool {
	return jerr.HasError(err, NotEnoughValueErrorText)
}

func IsNotEnoughTokenValueError(err error) bool {
	return jerr.HasError(err, NotEnoughTokenValueErrorText)
}
