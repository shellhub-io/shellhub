package dbtest

import (
	"path/filepath"
	"runtime"
)

func FixturesPath() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(file), "fixtures")
}
