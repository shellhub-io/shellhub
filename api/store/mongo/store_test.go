package mongo_test

import (
	"context"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/cache"
)

var srv = dbtest.DBServer{}
var s, _ = mongo.NewStore(context.TODO(), srv.Client().Database("test"), cache.NewNullCache())

func TestMain(m *testing.M) {
	fixtures.Init(srv.Host, "test")

	code := m.Run()
	srv.Stop()

	os.Exit(code)
}
