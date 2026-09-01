package dbtest

import (
	"path/filepath"
	"runtime"
)

// FixturesPath returns the fixtures directory, resolved from this file's own location so that
// tests find it whatever directory they run from.
func FixturesPath() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(file), "fixtures")
}
