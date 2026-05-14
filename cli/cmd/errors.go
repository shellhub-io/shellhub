package cmd

import (
	"errors"
)

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidUsername  = errors.New("username must be between 3 and 32 characters and can only contain lowercase letters, numbers, and -_.@")
	ErrInvalidPassword  = errors.New("password must be between 5 and 32 characters")
	ErrInvalidEmail     = errors.New("email is invalid")
	ErrInvalidNamespace = errors.New("namespace name is invalid")
	ErrInvalidType      = errors.New("namespace type must be either 'personal' or 'team'")
	ErrInvalidTenantID  = errors.New("tenant ID must be a valid UUID")
)
