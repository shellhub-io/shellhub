package store

import "github.com/shellhub-io/shellhub/pkg/errors"

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for store's error.
const ErrLayer = "store"

const (
	ErrCodeNoDocument = iota + 1
	ErrCodeDuplicated
	ErrCodeInvalid
)

var (
	ErrDuplicate   = errors.New("document duplicate", ErrLayer, ErrCodeDuplicated)
	ErrNoDocuments = errors.New("no documents", ErrLayer, ErrCodeNoDocument)
	ErrInvalidHex  = errors.New("the provided hex string is not a valid ObjectID", ErrLayer, ErrCodeInvalid)
)
