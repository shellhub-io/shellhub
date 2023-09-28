package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestSessionList(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		s     []models.Session
		count int
		err   error
	}

	cases := []struct {
		description string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when sessions are found",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session, fixtures.Device, fixtures.Namespace)
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
							CreatedAt:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							StatusUpdatedAt:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							LastSeen:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
							Name:             "hostname",
							Identity:         &models.DeviceIdentity{MAC: "mac"},
							Info:             nil,
							PublicKey:        "",
							TenantID:         "00000000-0000-4000-0000-000000000000",
							Online:           true,
							Namespace:        "namespace",
							Status:           "accepted",
							RemoteAddr:       "",
							Position:         nil,
							Tags:             []string{"tag1"},
							PublicURL:        false,
							PublicURLAddress: "",
							Acceptable:       false,
						},
						Active:        true,
						Closed:        false,
						Authenticated: true,
						Recorded:      false,
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			s, count, err := mongostore.SessionList(ctx, paginator.Query{Page: -1, PerPage: -1})
			assert.Equal(t, tc.expected, Expected{s: s, count: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionGet(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		s   *models.Session
		err error
	}

	cases := []struct {
		description string
		UID         models.UID
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when session is not found",
			UID:         models.UID("nonexistent"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session, fixtures.Device, fixtures.Namespace)
			},
			expected: Expected{
				s:   nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session, fixtures.Device, fixtures.Namespace)
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
						CreatedAt:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "hostname",
						Identity:         &models.DeviceIdentity{MAC: "mac"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           true,
						Namespace:        "namespace",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					Active:        true,
					Closed:        false,
					Authenticated: true,
					Recorded:      false,
					Type:          "",
					Term:          "",
					Position:      models.SessionPosition{Longitude: 0, Latitude: 0},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			s, err := mongostore.SessionGet(ctx, tc.UID)
			assert.Equal(t, tc.expected, Expected{s: s, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionCreate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		setup       func() error
		session     models.Session
		expected    error
	}{
		{
			description: "",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device, fixtures.Namespace)
			},
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			session, err := mongostore.SessionCreate(ctx, tc.session)
			assert.Equal(t, tc.expected, err)
			assert.NotEmpty(t, session)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionUpdateDeviceUID(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		oldUID      models.UID
		newUID      models.UID
		setup       func() error
		expected    error
	}{
		{
			description: "fails when device is not found",
			oldUID:      models.UID("nonexistent"),
			newUID:      models.UID("uid"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when device is found",
			oldUID:      models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			newUID:      models.UID("uid"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.SessionUpdateDeviceUID(ctx, tc.oldUID, tc.newUID)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionSetAuthenticated(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description  string
		UID          models.UID
		authenticate bool
		setup        func() error
		expected     error
	}{
		{
			description:  "fails when session is not found",
			UID:          models.UID("nonexistent"),
			authenticate: false,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description:  "succeeds when session is found",
			UID:          models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			authenticate: false,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.SessionSetAuthenticated(ctx, tc.UID, tc.authenticate)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionSetRecorded(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description  string
		UID          models.UID
		authenticate bool
		setup        func() error
		expected     error
	}{
		{
			description:  "fails when session is not found",
			UID:          models.UID("nonexistent"),
			authenticate: false,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description:  "succeeds when session is found",
			UID:          models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			authenticate: false,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.SessionSetAuthenticated(ctx, tc.UID, tc.authenticate)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionSetLastSeen(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		UID         models.UID
		setup       func() error
		expected    error
	}{
		{
			description: "fails when session is not found",
			UID:         models.UID("nonexistent"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.SessionSetLastSeen(ctx, tc.UID)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionDeleteActives(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		UID         models.UID
		setup       func() error
		expected    error
	}{
		{
			description: "fails when session is not found",
			UID:         models.UID("nonexistent"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.SessionDeleteActives(ctx, tc.UID)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionGetRecordFrame(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		r     []models.RecordedSession
		count int
		err   error
	}

	cases := []struct {
		description string
		UID         models.UID
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			r, count, err := mongostore.SessionGetRecordFrame(ctx, tc.UID)
			assert.Equal(t, tc.expected, Expected{r: r, count: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionCreateRecordFrame(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		UID         models.UID
		record      *models.RecordedSession
		setup       func() error
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
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
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
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.SessionCreateRecordFrame(ctx, tc.UID, tc.record)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestSessionDeleteRecordFrame(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		UID         models.UID
		setup       func() error
		expected    error
	}{
		{
			description: "fails when record frame is not found",
			UID:         models.UID("nonexistent"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when record frame is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.Session)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.SessionDeleteRecordFrame(ctx, tc.UID)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
