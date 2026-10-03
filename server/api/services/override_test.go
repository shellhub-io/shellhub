package services

import "testing"

func override[T any](t *testing.T, variable *T, value T) {
	t.Helper()

	previous := *variable
	t.Cleanup(func() { *variable = previous })
	*variable = value
}
