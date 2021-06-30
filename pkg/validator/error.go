package validator

import (
	"errors"
)

var (
	ErrStruct     = errors.New("the struct provided couldn't be checked")
	ErrBadRequest = errors.New("bad request")
)
