package mongo_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestSessionList(t *testing.T) {
	type Expected struct {
		s     []models.Session
		count int
		err   error
	}

	cases := []struct {
		description string
		paginator   query.Paginator
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when sessions are found",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureActiveSessions,
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
							CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
							Name:            "device-3",
							Identity:        &models.DeviceIdentity{MAC: "mac-3"},
							Info:            nil,
							PublicKey:       "",
							TenantID:        "00000000-0000-4000-0000-000000000000",
							Online:          false,
							Namespace:       "namespace-1",
							Status:          "accepted",
							RemoteAddr:      "",
							Position:        nil,
							Tags:            []string{"tag-1"},
							Acceptable:      false,
						},
						Active:        true,
						Closed:        true,
						Authenticated: true,
						Recorded:      false,
						Type:          "shell",
						Term:          "xterm",
						Position:      models.SessionPosition{Longitude: 0, Latitude: 0},
					},
					{
						StartedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:  time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:       "e7f3a56d8b9e1dc4c285c98c8ea9c33032a17bda5b6c6b05a6213c2a02f97824",
						DeviceUID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Username:  "john_doe",
						IPAddress: "0.0.0.0",
						Device: &models.Device{
							CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
							Name:            "device-3",
							Identity:        &models.DeviceIdentity{MAC: "mac-3"},
							Info:            nil,
							PublicKey:       "",
							TenantID:        "00000000-0000-4000-0000-000000000000",
							Online:          false,
							Namespace:       "namespace-1",
							Status:          "accepted",
							RemoteAddr:      "",
							Position:        nil,
							Tags:            []string{"tag-1"},
							Acceptable:      false,
						},
						Active:        false,
						Closed:        true,
						Authenticated: true,
						Recorded:      true,
						Type:          "shell",
						Term:          "xterm",
						Position:      models.SessionPosition{Longitude: 45.6789, Latitude: -12.3456},
					},
					{
						StartedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:       "fc2e1493d8b6a4c17bf6a2f7f9e55629e384b2d3a21e0c3d90f6e35b0c946178a",
						DeviceUID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Username:  "john_doe",
						IPAddress: "0.0.0.0",
						Device: &models.Device{
							CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
							Name:            "device-3",
							Identity:        &models.DeviceIdentity{MAC: "mac-3"},
							Info:            nil,
							PublicKey:       "",
							TenantID:        "00000000-0000-4000-0000-000000000000",
							Online:          false,
							Namespace:       "namespace-1",
							Status:          "accepted",
							RemoteAddr:      "",
							Position:        nil,
							Tags:            []string{"tag-1"},
							Acceptable:      false,
						},
						Active:        false,
						Closed:        true,
						Authenticated: true,
						Recorded:      false,
						Type:          "exec",
						Term:          "",
						Position:      models.SessionPosition{Longitude: -78.9012, Latitude: 23.4567},
					},
					{
						StartedAt: time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:       "bc3d75821a29cfe70bf7986f9ee5629e384b2d3a21e0c3d90f6e35b0c946178a",
						DeviceUID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Username:  "john_doe",
						IPAddress: "0.0.0.0",
						Device: &models.Device{
							CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
							UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
							Name:            "device-3",
							Identity:        &models.DeviceIdentity{MAC: "mac-3"},
							Info:            nil,
							PublicKey:       "",
							TenantID:        "00000000-0000-4000-0000-000000000000",
							Online:          false,
							Namespace:       "namespace-1",
							Status:          "accepted",
							RemoteAddr:      "",
							Position:        nil,
							Tags:            []string{"tag-1"},
							Acceptable:      false,
						},
						Active:        false,
						Closed:        true,
						Authenticated: true,
						Recorded:      true,
						Type:          "shell",
						Term:          "xterm",
						Position:      models.SessionPosition{Longitude: -56.7890, Latitude: 34.5678},
					},
				},
				count: 4,
				err:   nil,
			},
		},
	}

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(s []models.Session) {
		sort.Slice(s, func(i, j int) bool {
			return s[i].UID < s[j].UID
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			s, count, err := s.SessionList(ctx, tc.paginator)

			sort(tc.expected.s)
			sort(s)

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
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureActiveSessions,
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
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureActiveSessions,
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
						CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:            "device-3",
						Identity:        &models.DeviceIdentity{MAC: "mac-3"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					Active:        true,
					Closed:        true,
					Authenticated: true,
					Recorded:      false,
					Type:          "shell",
					Term:          "xterm",
					Position:      models.SessionPosition{Longitude: 0, Latitude: 0},
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

			s, err := s.SessionGet(ctx, tc.UID)
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
			fixtures:    []string{fixtureDevices, fixtureNamespaces},
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
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			session, err := s.SessionCreate(ctx, tc.session)
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
			fixtures:    []string{fixtureSessions},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when device is found",
			oldUID:      models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			newUID:      models.UID("uid"),
			fixtures:    []string{fixtureSessions},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.SessionUpdateDeviceUID(ctx, tc.oldUID, tc.newUID)
			assert.Equal(t, tc.expected, err)
		})
	}
}

// ptrBool and ptrString are helpers for creating pointer values in tests.
func ptrBool(b bool) *bool       { return &b }
func ptrString(s string) *string { return &s }

// TestSessionUpdate exercises different update paths for the SessionUpdate method.
func TestSessionUpdate(t *testing.T) {
	type args struct {
		sess   *models.Session
		update *models.SessionUpdate
	}
	cases := []struct {
		description string
		UID         models.UID
		args        args
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when session is found and no update fields",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			args: args{
				sess:   &models.Session{Authenticated: true},
				update: &models.SessionUpdate{},
			},
			fixtures: []string{fixtureSessions},
			expected: nil,
		},
		{
			description: "succeeds when setting Authenticated to true",
			UID:         models.UID("e7f3a56d8b9e1dc4c285c98c8ea9c33032a17bda5b6c6b05a6213c2a02f97824"),
			args: args{
				sess:   &models.Session{Authenticated: false, StartedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC), TenantID: "00000000-0000-4000-0000-000000000000"},
				update: &models.SessionUpdate{Authenticated: ptrBool(true)},
			},
			fixtures: []string{fixtureSessions},
			expected: nil,
		},
		{
			description: "succeeds when updating Type field",
			UID:         models.UID("fc2e1493d8b6a4c17bf6a2f7f9e55629e384b2d3a21e0c3d90f6e35b0c946178a"),
			args: args{
				sess:   &models.Session{},
				update: &models.SessionUpdate{Type: ptrString("exec")},
			},
			fixtures: []string{fixtureSessions},
			expected: nil,
		},
		{
			description: "succeeds when updating Recorded flag",
			UID:         models.UID("bc3d75821a29cfe70bf7986f9ee5629e384b2d3a21e0c3d90f6e35b0c946178a"),
			args: args{
				sess:   &models.Session{},
				update: &models.SessionUpdate{Recorded: ptrBool(true)},
			},
			fixtures: []string{fixtureSessions},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.SessionUpdate(ctx, tc.UID, tc.args.sess, tc.args.update)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionSetRecorded(t *testing.T) {
	cases := []struct {
		description string
		UID         models.UID
		recorded    bool
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when session is not found",
			UID:         models.UID("nonexistent"),
			recorded:    false,
			fixtures:    []string{fixtureSessions},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			recorded:    false,
			fixtures:    []string{fixtureSessions},
			expected:    nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})
			err := s.SessionSetRecorded(ctx, tc.UID, tc.recorded)
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
			fixtures:    []string{fixtureSessions},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			fixtures:    []string{fixtureSessions},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.SessionSetLastSeen(ctx, tc.UID)
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
			fixtures:    []string{fixtureSessions},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when session is found",
			UID:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			fixtures:    []string{fixtureSessions},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.SessionDeleteActives(ctx, tc.UID)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestSessionListEvents(t *testing.T) {
	type Expected struct {
		events []models.SessionEvent
		count  int
		err    error
	}

	cases := []struct {
		description string
		uid         string
		paginator   query.Paginator
		sorter      query.Sorter
		filters     query.Filters
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when sessions are not found",
			uid:         "nonexistent",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			sorter:      query.Sorter{By: "timestamp", Order: query.OrderAsc},
			filters:     query.Filters{},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureActiveSessions,
				fixtureSessionsEvents,
			},
			expected: Expected{
				events: []models.SessionEvent{},
				count:  0,
				err:    nil,
			},
		},
		{
			description: "succeeds when sessions are found",
			uid:         "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			sorter:      query.Sorter{By: "timestamp", Order: query.OrderAsc},
			filters:     query.Filters{},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureActiveSessions,
				fixtureSessionsEvents,
			},
			expected: Expected{
				events: []models.SessionEvent{
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "pty-req",
						Timestamp: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						Data: models.SSHPty{
							Term:     "screen-256color",
							Columns:  211,
							Rows:     47,
							Width:    1899,
							Height:   940,
							Modelist: []byte{},
						},
						Seat: 0,
					},
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "shell",
						Timestamp: time.Date(2023, 1, 2, 12, 1, 0, 0, time.UTC),
						Data:      "",
						Seat:      0,
					},
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "exit-status",
						Timestamp: time.Date(2023, 1, 2, 12, 2, 0, 0, time.UTC),
						Data:      "AAAAAA==",
						Seat:      0,
					},
				},
				count: 3,
				err:   nil,
			},
		},
		{
			description: "succeeds when sessions are found by page are limited",
			uid:         "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
			paginator:   query.Paginator{Page: 1, PerPage: 2},
			sorter:      query.Sorter{By: "timestamp", Order: query.OrderAsc},
			filters:     query.Filters{},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureActiveSessions,
				fixtureSessionsEvents,
			},
			expected: Expected{
				events: []models.SessionEvent{
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "pty-req",
						Timestamp: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						Data: models.SSHPty{
							Term:     "screen-256color",
							Columns:  211,
							Rows:     47,
							Width:    1899,
							Height:   940,
							Modelist: []byte{},
						},
						Seat: 0,
					},
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "shell",
						Timestamp: time.Date(2023, 1, 2, 12, 1, 0, 0, time.UTC),
						Data:      "",
						Seat:      0,
					},
				},
				count: 3,
				err:   nil,
			},
		},
		{
			description: "succeeds when filtering by event type",
			uid:         "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "timestamp", Order: query.OrderAsc},
			filters: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "type",
							Operator: "eq",
							Value:    "pty-req",
						},
					},
				},
			},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureActiveSessions,
				fixtureSessionsEvents,
			},
			expected: Expected{
				events: []models.SessionEvent{
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "pty-req",
						Timestamp: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						Data: models.SSHPty{
							Term:     "screen-256color",
							Columns:  211,
							Rows:     47,
							Width:    1899,
							Height:   940,
							Modelist: []byte{},
						},
						Seat: 0,
					},
				},
				count: 1,
				err:   nil,
			},
		},
		{
			description: "succeeds when filtering by seat",
			uid:         "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "timestamp", Order: query.OrderAsc},
			filters: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "seat",
							Operator: "eq",
							Value:    0, // Use integer instead of string
						},
					},
				},
			},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureActiveSessions,
				fixtureSessionsEvents,
			},
			expected: Expected{
				events: []models.SessionEvent{
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "pty-req",
						Timestamp: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						Data: models.SSHPty{
							Term:     "screen-256color",
							Columns:  211,
							Rows:     47,
							Width:    1899,
							Height:   940,
							Modelist: []byte{},
						},
						Seat: 0,
					},
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "shell",
						Timestamp: time.Date(2023, 1, 2, 12, 1, 0, 0, time.UTC),
						Data:      "",
						Seat:      0,
					},
					{
						Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
						Type:      "exit-status",
						Timestamp: time.Date(2023, 1, 2, 12, 2, 0, 0, time.UTC),
						Data:      "AAAAAA==",
						Seat:      0,
					},
				},
				count: 3,
				err:   nil,
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

			events, count, err := s.SessionListEvents(ctx, models.UID(tc.uid), tc.paginator, tc.filters, tc.sorter)

			assert.Equal(t, tc.expected, Expected{events: events, count: count, err: err})
		})
	}
}

func TestSessionEvent(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		event       *models.SessionEvent
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when creating a new session event",
			uid:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			event: &models.SessionEvent{
				Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
				Type:      models.SessionEventTypePtyRequest,
				Timestamp: time.Date(2023, 1, 2, 12, 3, 0, 0, time.UTC),
				Data: models.SSHPty{
					Term:     "xterm-256color",
					Columns:  80,
					Rows:     24,
					Width:    640,
					Height:   480,
					Modelist: []byte{},
				},
				Seat: 0,
			},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
			},
			expected: nil,
		},
		{
			description: "succeeds when creating a window change event",
			uid:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			event: &models.SessionEvent{
				Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
				Type:      models.SessionEventTypeWindowChange,
				Timestamp: time.Date(2023, 1, 2, 12, 4, 0, 0, time.UTC),
				Data: models.SSHWindowChange{
					Columns: 120,
					Rows:    30,
					Width:   960,
					Height:  720,
				},
				Seat: 0,
			},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
			},
			expected: nil,
		},
		{
			description: "succeeds when creating an exit status event",
			uid:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			event: &models.SessionEvent{
				Session:   "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
				Type:      models.SessionEventTypeExitStatus,
				Timestamp: time.Date(2023, 1, 2, 12, 5, 0, 0, time.UTC),
				Data:      "0",
				Seat:      0,
			},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
			},
			expected: nil,
		},
		{
			description: "succeeds when session does not exist",
			uid:         models.UID("nonexistent"),
			event: &models.SessionEvent{
				Session:   "nonexistent",
				Type:      models.SessionEventTypePtyRequest,
				Timestamp: time.Date(2023, 1, 2, 12, 3, 0, 0, time.UTC),
				Data:      "",
				Seat:      0,
			},
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.SessionEvent(ctx, tc.uid, tc.event)
			assert.Equal(t, tc.expected, err)

			// Verify the event was created in sessions_events collection
			if tc.expected == nil {
				var event models.SessionEvent
				store := s.(*mongo.Store)
				err := store.GetDB().Collection("sessions_events").FindOne(ctx, bson.M{
					"session":   tc.event.Session,
					"type":      tc.event.Type,
					"timestamp": tc.event.Timestamp,
					"seat":      tc.event.Seat,
				}).Decode(&event)
				assert.NoError(t, err)
				assert.Equal(t, tc.event.Session, event.Session)
				assert.Equal(t, tc.event.Type, event.Type)
				assert.Equal(t, tc.event.Seat, event.Seat)
			}
		})
	}
}

func TestSessionDeleteEvents(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		seat        int
		eventType   models.SessionEventType
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when deleting existing events",
			uid:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			seat:        0,
			eventType:   models.SessionEventTypePtyRequest,
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureSessionsEvents,
			},
			expected: nil,
		},
		{
			description: "succeeds when deleting shell events",
			uid:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			seat:        0,
			eventType:   models.SessionEventTypeShell,
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureSessionsEvents,
			},
			expected: nil,
		},
		{
			description: "succeeds when deleting exit status events",
			uid:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			seat:        0,
			eventType:   models.SessionEventTypeExitStatus,
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureSessionsEvents,
			},
			expected: nil,
		},
		{
			description: "succeeds when no events match criteria",
			uid:         models.UID("a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68"),
			seat:        1,
			eventType:   models.SessionEventTypePtyRequest,
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureSessionsEvents,
			},
			expected: nil,
		},
		{
			description: "succeeds when session does not exist",
			uid:         models.UID("nonexistent"),
			seat:        0,
			eventType:   models.SessionEventTypePtyRequest,
			fixtures: []string{
				fixtureNamespaces,
				fixtureDevices,
				fixtureSessions,
				fixtureSessionsEvents,
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			store := s.(*mongo.Store)
			countBefore, err := store.GetDB().Collection("sessions_events").CountDocuments(ctx, bson.M{
				"session": tc.uid,
				"seat":    tc.seat,
				"type":    tc.eventType,
			})
			assert.NoError(t, err)

			totalBefore, err := store.GetDB().Collection("sessions_events").CountDocuments(ctx, bson.M{})
			assert.NoError(t, err)

			err = s.SessionDeleteEvents(ctx, tc.uid, tc.seat, tc.eventType)
			assert.Equal(t, tc.expected, err)

			countAfter, err := store.GetDB().Collection("sessions_events").CountDocuments(ctx, bson.M{
				"session": tc.uid,
				"seat":    tc.seat,
				"type":    tc.eventType,
			})
			assert.NoError(t, err)
			assert.Equal(t, int64(0), countAfter)

			totalAfter, err := store.GetDB().Collection("sessions_events").CountDocuments(ctx, bson.M{})
			assert.NoError(t, err)

			assert.Equal(t, totalBefore-countBefore, totalAfter)
		})
	}
}
