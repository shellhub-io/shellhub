package migrations

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("SHELLHUB_ENTERPRISE", "true")
	os.Setenv("SHELLHUB_CLOUD", "true")
	code := m.Run()
	os.Exit(code)
}
