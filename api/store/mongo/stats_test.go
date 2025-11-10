package mongo_test

import (
	"context"
	"testing"

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
		tenantID    string
		expected    Expected
	}{
		{
			description: "succeeds without tenantID",
			fixtures:    []string{fixtureUsers, fixtureNamespaces, fixtureSessions, fixtureActiveSessions, fixtureDevices},
			tenantID:    "",
			expected: Expected{
				stats: &models.Stats{
					RegisteredDevices: 3,
					OnlineDevices:     0,
					ActiveSessions:    1,
					PendingDevices:    1,
					RejectedDevices:   0,
				},
				err: nil,
			},
		},
		{
			description: "succeeds with specific tenantID",
			fixtures:    []string{fixtureUsers, fixtureNamespaces, fixtureSessions, fixtureActiveSessions, fixtureDevices},
			tenantID:    "00000000-0000-4000-0000-000000000000",
			expected: Expected{
				stats: &models.Stats{
					RegisteredDevices: 3,
					OnlineDevices:     0,
					ActiveSessions:    1,
					PendingDevices:    1,
					RejectedDevices:   0,
				},
				err: nil,
			},
		},
		{
			description: "succeeds with non-existent tenantID",
			fixtures:    []string{fixtureUsers, fixtureNamespaces, fixtureSessions, fixtureActiveSessions, fixtureDevices},
			tenantID:    "99999999-9999-4999-9999-999999999999",
			expected: Expected{
				stats: &models.Stats{
					RegisteredDevices: 0,
					OnlineDevices:     0,
					ActiveSessions:    0,
					PendingDevices:    0,
					RejectedDevices:   0,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			stats, err := s.GetStats(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{stats: stats, err: err})
		})
	}
}
