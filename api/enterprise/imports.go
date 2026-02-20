//go:build cloud

package enterprise

import (
	// Blank import triggers the init() functions of all cloud packages via the
	// register entry point, registering billing providers, route extensions,
	// and worker extensions with the core API before the server initializes.
	_ "github.com/shellhub-io/cloud/enterprise"
)
