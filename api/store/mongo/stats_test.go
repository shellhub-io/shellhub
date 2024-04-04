package mongo_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGetStats(t *testing.T) {
	type Expected struct {
		stats *models.Stats
		err   error
	}

	cases := []struct {
		description string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds",
			fixtures: []string{
				fixtures.FixtureUsers,
				fixtures.FixtureNamespaces,
				fixtures.FixtureSessions,
				fixtures.FixtureActiveSessions,
				fixtures.FixtureDevices,
				fixtures.FixtureConnectedDevices,
			},
			expected: Expected{
				stats: &models.Stats{
					RegisteredDevices: 3,
					OnlineDevices:     1,
					ActiveSessions:    1,
					PendingDevices:    1,
					RejectedDevices:   0,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			stats, err := store.GetStats(ctx)
			assert.Equal(t, tc.expected, Expected{stats: stats, err: err})
		})
	}
}
