package store

import (
	"errors"
	"fmt"
)

func newError(err error) error {
	return fmt.Errorf("store: %w", err)
}

var (
	ErrNamespaceRename   = newError(errors.New("could not rename the namespace"))
	ErrNamespaceNotFound = newError(errors.New("not found"))
)

var (
	ErrDuplicate   = errors.New("duplicate")
	ErrNoDocuments = errors.New("mongo: no documents in result")
	ErrInvalidHex  = errors.New("the provided hex string is not a valid ObjectID")
)
