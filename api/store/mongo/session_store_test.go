package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestSessionList(t *testing.T) {
	type Expected struct {
		s     []models.Session
		count int
		err   error
	}

	cases := []struct {
		description string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when sessions are found",
			fixtures: []string{
				fixtures.FixtureNamespaces,
				fixtures.FixtureDevices,
				fixtures.FixtureConnectedDevices,
				fixtures.FixtureSessions,
				fixtures.FixtureActiveSessions,
			},
			expected: Expected{
				s: []models.Session{
					{
						StartedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:       "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						DeviceUID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Username:  "john_doe",
						IPAddress: "0.0.0.0",
						Device: &models.Device{
							CreatedAt:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							StatusUpdatedAt:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							LastSeen:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
							Name:             "device-3",
							Identity:         &models.DeviceIdentity{MAC: "mac-3"},
							Info:             nil,
							PublicKey:        "",
							TenantID:         "00000000-0000-4000-0000-000000000000",
							Online:           true,
							Namespace:        "namespace-1",
							Status:           "accepted",
							RemoteAddr:       "",
							Position:         nil,
							Tags:             []string{"tag-1"},
							PublicURL:        false,
							PublicURLAddress: "",
							Acceptable:       false,
						},
						Active:        true,
						Closed:        false,
						Authenticated: true,
						Recorded:      true,
						Type:          "",
						Term:          "",
						Position:      models.SessionPosition{Longitude: 0, Latitude: 0},
					},
				},
				count: 1,
				err:   nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			s, count, err := mongostore.SessionList(context.TODO(), paginator.Query{Page: -1, PerPage: -1})
			assert.Equal(t, tc.expected, Expected{s: s, count: count, err: err})
		})
	}
}

func TestSessionGet(t *testing.T) {
	type Expected struct {
		s   *models.Session
		err error
	}

	cases := []struct {
		description string
		UID         models.UID
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when session is not found",
			UID:         models.UID("nonexistent"),
			fixtures: []string{
				fixtures.FixtureNamespaces,
				fixtures.FixtureDevices,
				fixtures.FixtureConnectedDevices,
				fixtures.FixtureSessions,
				fixtures.FixtureActiveSessions,
			},
			expected: Expected{
				s:   nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			fixtures: []string{
				fixtures.FixtureNamespaces,
				fixtures.FixtureDevices,
				fixtures.FixtureConnectedDevices,
				fixtures.FixtureSessions,
				fixtures.FixtureActiveSessions,
			},
			expected: Expected{
				s: &models.Session{
					StartedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastSeen:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UID:       "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
					DeviceUID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Username:  "john_doe",
					IPAddress: "0.0.0.0",
					Device: &models.Device{
						CreatedAt:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "device-3",
						Identity:         &models.DeviceIdentity{MAC: "mac-3"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           true,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					Active:        true,
					Closed:        false,
					Authenticated: true,
					Recorded:      true,
					Type:          "",
					Term:          "",
					Position:      models.SessionPosition{Longitude: 0, Latitude: 0},
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			s, err := mongostore.SessionGet(context.TODO(), tc.UID)
			assert.Equal(t, tc.expected, Expected{s: s, err: err})
		})
	}
}

func TestSessionCreate(t *testing.T) {
	cases := []struct {
		description string
		fixtures    []string
		session     models.Session
		expected    error
	}{
		{
			description: "",
			fixtures:    []string{fixtures.FixtureDevices, fixtures.FixtureNamespaces},
			session: models.Session{
				Username:      "username",
				UID:           "uid",
				TenantID:      "00000000-0000-4000-0000-000000000000",
				DeviceUID:     models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
				IPAddress:     "0.0.0.0",
				Authenticated: true,
			},
			expected: nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			session, err := mongostore.SessionCreate(context.TODO(), tc.session)
			assert.Equal(t, tc.expected, err)
			assert.NotEmpty(t, session)
		})
	}
}

func TestSessionUpdateDeviceUID(t *testing.T) {
	cases := []struct {
		description string
		oldUID      models.UID
		newUID      models.UID
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when device is not found",
			oldUID:      models.UID("nonexistent"),
			newUID:      models.UID("uid"),
			fixtures:    []string{fixtures.FixtureSessions},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when device is found",
			oldUID:      models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			newUID:      models.UID("uid"),
			fixtures:    []string{fixtures.FixtureSessions},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.SessionUpdateDeviceUID(context.TODO(), tc.oldUID, tc.newUID)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionSetAuthenticated(t *testing.T) {
	cases := []struct {
		description  string
		UID          models.UID
		authenticate bool
		fixtures     []string
		expected     error
	}{
		{
			description:  "fails when session is not found",
			UID:          models.UID("nonexistent"),
			authenticate: false,
			fixtures:     []string{fixtures.FixtureSessions},
			expected:     store.ErrNoDocuments,
		},
		{
			description:  "succeeds when session is found",
			UID:          models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			authenticate: false,
			fixtures:     []string{fixtures.FixtureSessions},
			expected:     nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.SessionSetAuthenticated(context.TODO(), tc.UID, tc.authenticate)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionSetRecorded(t *testing.T) {
	cases := []struct {
		description  string
		UID          models.UID
		authenticate bool
		fixtures     []string
		expected     error
	}{
		{
			description:  "fails when session is not found",
			UID:          models.UID("nonexistent"),
			authenticate: false,
			fixtures:     []string{fixtures.FixtureSessions},
			expected:     store.ErrNoDocuments,
		},
		{
			description:  "succeeds when session is found",
			UID:          models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			authenticate: false,
			fixtures:     []string{fixtures.FixtureSessions},
			expected:     nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.SessionSetAuthenticated(context.TODO(), tc.UID, tc.authenticate)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionSetLastSeen(t *testing.T) {
	cases := []struct {
		description string
		UID         models.UID
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when session is not found",
			UID:         models.UID("nonexistent"),
			fixtures:    []string{fixtures.FixtureSessions},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			fixtures:    []string{fixtures.FixtureSessions},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.SessionSetLastSeen(context.TODO(), tc.UID)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionDeleteActives(t *testing.T) {
	cases := []struct {
		description string
		UID         models.UID
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when session is not found",
			UID:         models.UID("nonexistent"),
			fixtures:    []string{fixtures.FixtureSessions},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			fixtures:    []string{fixtures.FixtureSessions},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.SessionDeleteActives(context.TODO(), tc.UID)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionGetRecordFrame(t *testing.T) {
	type Expected struct {
		r     []models.RecordedSession
		count int
		err   error
	}

	cases := []struct {
		description string
		UID         models.UID
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			fixtures:    []string{fixtures.FixtureSessions, fixtures.FixtureRecordedSessions},
			expected: Expected{
				r: []models.RecordedSession{
					{
						Time:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:      "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Message:  "message",
						TenantID: "00000000-0000-4000-0000-000000000000",
						Width:    0,
						Height:   0,
					},
				},
				count: 1,
				err:   nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			r, count, err := mongostore.SessionGetRecordFrame(context.TODO(), tc.UID)
			assert.Equal(t, tc.expected, Expected{r: r, count: count, err: err})
		})
	}
}

func TestSessionCreateRecordFrame(t *testing.T) {
	cases := []struct {
		description string
		UID         models.UID
		record      *models.RecordedSession
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when session is not found",
			UID:         models.UID("nonexistent"),
			record: &models.RecordedSession{
				UID:      models.UID("nonexistent"),
				Message:  "message",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Time:     time.Now(),
				Width:    0,
				Height:   0,
			},
			fixtures: []string{fixtures.FixtureSessions},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			record: &models.RecordedSession{
				UID:      models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
				Message:  "message",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Time:     time.Now(),
				Width:    0,
				Height:   0,
			},
			fixtures: []string{fixtures.FixtureSessions},
			expected: nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.SessionCreateRecordFrame(context.TODO(), tc.UID, tc.record)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionDeleteRecordFrame(t *testing.T) {
	cases := []struct {
		description string
		UID         models.UID
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when record frame is not found",
			UID:         models.UID("nonexistent"),
			fixtures:    []string{fixtures.FixtureSessions, fixtures.FixtureRecordedSessions},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when record frame is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			fixtures:    []string{fixtures.FixtureSessions, fixtures.FixtureRecordedSessions},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.SessionDeleteRecordFrame(context.TODO(), tc.UID)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionDeleteRecordFrameByDate(t *testing.T) {
	type Expected struct {
		deletedCount int64
		updatedCount int64
		err          error
	}

	cases := []struct {
		description string
		lte         time.Time
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when there are no sessions to update or delete",
			lte:         time.Date(2023, time.February, 10, 8, 30, 0, 0, time.UTC),
			fixtures:    []string{},
			expected: Expected{
				deletedCount: 0,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "succeeds to delete and update recorded sessions before specified date",
			lte:         time.Date(2023, time.January, 2, 12, 01, 0, 0, time.UTC),
			fixtures: []string{
				fixtures.FixtureSessions,
				fixtures.FixtureRecordedSessions,
			},
			expected: Expected{
				deletedCount: 1,
				updatedCount: 1,
				err:          nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint:errcheck

			deletedCount, updatedCount, err := mongostore.SessionDeleteRecordFrameByDate(context.TODO(), tc.lte)
			assert.Equal(t, tc.expected, Expected{deletedCount, updatedCount, err})
		})
	}
}
