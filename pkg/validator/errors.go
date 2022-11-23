package validator

import (
	"errors"
)

var ErrInvalidError = errors.New("this error is not from a field validation")
