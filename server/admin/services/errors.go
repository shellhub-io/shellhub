package services

import (
	"errors"
)

var (
	ErrCreateNewUser               = errors.New("failed to create a new user")
	ErrDuplicateNamespace          = errors.New("namespace already exists")
	ErrNamespaceSingle             = errors.New("this instance does not support multi-tenancy")
	ErrNamespaceInstanceProtected  = errors.New("this namespace is bound to the instance and cannot be removed")
	ErrUserNotFound                = errors.New("user not found")
	ErrNamespaceNotFound           = errors.New("namespace not found")
	ErrFailedDeleteUser            = errors.New("failed to delete the user")
	ErrFailedDeleteNamespace       = errors.New("failed to delete the namespace")
	ErrFailedUpdateUser            = errors.New("failed to reset the password for the user")
	ErrFailedNamespaceRemoveMember = errors.New("failed to remove member from the namespace")
	ErrUserPasswordInvalid         = errors.New("user password is invalid")
	ErrUserEmailExists             = errors.New("user email already exists")
	ErrUserNameExists              = errors.New("user name already exists")
	ErrFailedNamespaceAddMember    = errors.New("could not add this member to this namespace")
	ErrUserUnhandledDuplicate      = errors.New("unhandled duplicated field for the user")
	ErrFailedListNamespaces        = errors.New("failed to list namespaces")
	ErrFailedListUsers             = errors.New("failed to list users")
)
