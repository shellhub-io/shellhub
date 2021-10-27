package migrations

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("SHELLHUB_ENTERPRISE", "true")
	code := m.Run()
	os.Exit(code)
}
