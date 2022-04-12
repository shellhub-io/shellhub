package guard

import "github.com/shellhub-io/shellhub/pkg/errors"

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for guard's error.
var ErrLayer = "guard"

// ErrCodeForbidden is the error code when the access to a resource is forbidden.
const ErrCodeForbidden = iota + 1

// ErrForbidden is used to indicate that access to a resource is forbidden.
var ErrForbidden = errors.New("access forbidden", ErrLayer, ErrCodeForbidden)
