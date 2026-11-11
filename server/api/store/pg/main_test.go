package pg_test

import (
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Run(m))
}
