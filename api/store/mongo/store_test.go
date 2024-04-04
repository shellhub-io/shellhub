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

var db = dbtest.DBServer{}
var store, _ = mongo.NewStore(context.TODO(), db.Client().Database("test"), cache.NewNullCache())

func TestMain(m *testing.M) {
	defer db.Stop()
	fixtures.Init(db.Host, "test")

	code := m.Run()
	os.Exit(code)
}
