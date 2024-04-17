package migrations

import (
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
)

var srv = dbtest.DBServer{}

func TestMain(m *testing.M) {
	os.Setenv("SHELLHUB_ENTERPRISE", "true")
	os.Setenv("SHELLHUB_CLOUD", "true")

	_ = srv.Client() // fills srv.Host
	fixtures.Init(srv.Host, "test")

	code := m.Run()
	srv.Stop()

	os.Exit(code)
}
