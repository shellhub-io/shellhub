package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGetStats(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		stats *models.Stats
		err   error
	}

	cases := []struct {
		description string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User, fixtures.Namespace, fixtures.Session, fixtures.Device)
			},
			expected: Expected{
				stats: &models.Stats{
					RegisteredDevices: 1,
					OnlineDevices:     1,
					ActiveSessions:    1,
					PendingDevices:    0,
					RejectedDevices:   0,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			stats, err := mongostore.GetStats(ctx)
			assert.Equal(t, tc.expected, Expected{stats: stats, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
