package osauth

import "errors"

// ErrUserNotFound is returned when the user is not found in the passwd file.
var ErrUserNotFound = errors.New("user not found")

var (
	errMalformedUID = errors.New("malformed uid field")
	errMalformedGID = errors.New("malformed gid field")
)
