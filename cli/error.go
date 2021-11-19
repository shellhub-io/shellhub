package main

import (
	"errors"
)

var (
	ErrCreateNewUser               = errors.New("failed to create a new user")
	ErrDuplicateNamespace          = errors.New("namespace already exists")
	ErrUserNotFound                = errors.New("user not found")
	ErrNamespaceNotFound           = errors.New("namespace not found")
	ErrFailedDeleteUser            = errors.New("failed to delete the user")
	ErrFailedDeleteNamespace       = errors.New("failed to delete the namespace")
	ErrFailedUpdateUser            = errors.New("failed to reset the password for the user")
	ErrFailedNamespaceRemoveMember = errors.New("failed to remove member from the namespace")
	ErrUserDataInvalid             = errors.New("user data is invalid")
	ErrUserPasswordInvalid         = errors.New("user password is invalid")
	ErrUserEmailExists             = errors.New("user email already exists")
	ErrUserNameExists              = errors.New("user name already exists")
	ErrUserNameAndEmailExists      = errors.New("user name and email already exists")
	ErrPasswordInvalid             = errors.New("password is invalid")
	ErrNamespaceInvalid            = errors.New("namespace is invalid")
	ErrFailedNamespaceAddMember    = errors.New("could not add this member to this namespace")
)
