package mongo_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
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
		opts        []store.QueryOption
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when sessions are found",
			opts:        []store.QueryOption{s.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1})},
			fixtures: []string{
				fixtureNamespaces,
				fixtureTags,
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
							Acceptable:      false,
							Taggable: models.Taggable{
								TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
								Tags: []models.Tag{
									{
										ID:        "6791d3ae04ba86e6d7a0514d",
										CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										Name:      "production",
										TenantID:  "00000000-0000-4000-0000-000000000000",
									},
									{
										ID:        "6791d3be5a201d874c4c2885",
										CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										Name:      "development",
										TenantID:  "00000000-0000-4000-0000-000000000000",
									},
								},
							},
						},
						Active:        true,
						Closed:        true,
						Authenticated: true,
						Recorded:      false,
						Type:          "shell",
						Term:          "xterm",
						Position:      models.SessionPosition{Longitude: 0, Latitude: 0},
						Events:        models.SessionEvents{Types: []string{}, Seats: []int{}},
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
							Acceptable:      false,
							Taggable: models.Taggable{
								TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
								Tags: []models.Tag{
									{
										ID:        "6791d3ae04ba86e6d7a0514d",
										CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										Name:      "production",
										TenantID:  "00000000-0000-4000-0000-000000000000",
									},
									{
										ID:        "6791d3be5a201d874c4c2885",
										CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										Name:      "development",
										TenantID:  "00000000-0000-4000-0000-000000000000",
									},
								},
							},
						},
						Active:        false,
						Closed:        true,
						Authenticated: true,
						Recorded:      true,
						Type:          "shell",
						Term:          "xterm",
						Position:      models.SessionPosition{Longitude: 45.6789, Latitude: -12.3456},
						Events:        models.SessionEvents{Types: []string{}, Seats: []int{}},
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
							Acceptable:      false,
							Taggable: models.Taggable{
								TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
								Tags: []models.Tag{
									{
										ID:        "6791d3ae04ba86e6d7a0514d",
										CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										Name:      "production",
										TenantID:  "00000000-0000-4000-0000-000000000000",
									},
									{
										ID:        "6791d3be5a201d874c4c2885",
										CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										Name:      "development",
										TenantID:  "00000000-0000-4000-0000-000000000000",
									},
								},
							},
						},
						Active:        false,
						Closed:        true,
						Authenticated: true,
						Recorded:      false,
						Type:          "exec",
						Term:          "",
						Position:      models.SessionPosition{Longitude: -78.9012, Latitude: 23.4567},
						Events:        models.SessionEvents{Types: []string{}, Seats: []int{}},
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
							Acceptable:      false,
							Taggable: models.Taggable{
								TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
								Tags: []models.Tag{
									{
										ID:        "6791d3ae04ba86e6d7a0514d",
										CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										Name:      "production",
										TenantID:  "00000000-0000-4000-0000-000000000000",
									},
									{
										ID:        "6791d3be5a201d874c4c2885",
										CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
										Name:      "development",
										TenantID:  "00000000-0000-4000-0000-000000000000",
									},
								},
							},
						},
						Active:        false,
						Closed:        true,
						Authenticated: true,
						Recorded:      true,
						Type:          "shell",
						Term:          "xterm",
						Position:      models.SessionPosition{Longitude: -56.7890, Latitude: 34.5678},
						Events:        models.SessionEvents{Types: []string{}, Seats: []int{}},
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

			s, count, err := s.SessionList(ctx, tc.opts...)

			sort(tc.expected.s)
			sort(s)

			assert.Equal(t, tc.expected, Expected{s: s, count: count, err: err})
		})
	}
}

func TestSessionResolve(t *testing.T) {
	type Expected struct {
		s   *models.Session
		err error
	}

	cases := []struct {
		description string
		resolver    store.SessionResolver
		value       string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when session is not found",
			resolver:    store.SessionUIDResolver,
			value:       "nonexistent",
			fixtures: []string{
				fixtureNamespaces,
				fixtureTags,
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
			resolver:    store.SessionUIDResolver,
			value:       "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
			fixtures: []string{
				fixtureNamespaces,
				fixtureTags,
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
							Tags: []models.Tag{
								{
									ID:        "6791d3ae04ba86e6d7a0514d",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "production",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
								{
									ID:        "6791d3be5a201d874c4c2885",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "development",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
							},
						},
					},
					Active:        true,
					Closed:        true,
					Authenticated: true,
					Recorded:      false,
					Type:          "shell",
					Term:          "xterm",
					Position:      models.SessionPosition{Longitude: 0, Latitude: 0},
					Events:        models.SessionEvents{Types: []string{}, Seats: []int{}},
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

			s, err := s.SessionResolve(ctx, tc.resolver, tc.value)
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

			uid, err := s.SessionCreate(ctx, tc.session)
			assert.Equal(t, tc.expected, err)
			assert.NotEmpty(t, uid)
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

// TestSessionUpdate exercises different update paths for the SessionUpdate method.
func TestSessionUpdate(t *testing.T) {
	cases := []struct {
		description string
		session     *models.Session
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when session is found",
			session:     &models.Session{UID: "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68", Authenticated: true},
			fixtures:    []string{fixtureSessions},
			expected:    nil,
		},
		{
			description: "succeeds when setting Authenticated to true",
			session:     &models.Session{UID: "e7f3a56d8b9e1dc4c285c98c8ea9c33032a17bda5b6c6b05a6213c2a02f97824", Authenticated: true, StartedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC), TenantID: "00000000-0000-4000-0000-000000000000"},
			fixtures:    []string{fixtureSessions},
			expected:    nil,
		},
		{
			description: "succeeds when updating Type field",
			session:     &models.Session{UID: "fc2e1493d8b6a4c17bf6a2f7f9e55629e384b2d3a21e0c3d90f6e35b0c946178a", Type: "exec"},
			fixtures:    []string{fixtureSessions},
			expected:    nil,
		},
		{
			description: "succeeds when updating Recorded flag",
			session:     &models.Session{UID: "bc3d75821a29cfe70bf7986f9ee5629e384b2d3a21e0c3d90f6e35b0c946178a", Recorded: true},
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

			err := s.SessionUpdate(ctx, tc.session)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestActiveSessionDelete(t *testing.T) {
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

			err := s.ActiveSessionDelete(ctx, tc.UID)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestActiveSessionResolve(t *testing.T) {
	type Expected struct {
		activeSession *models.ActiveSession
		err           error
	}

	cases := []struct {
		description string
		resolver    store.SessionResolver
		value       string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when active session is not found",
			resolver:    store.SessionUIDResolver,
			value:       "nonexistent",
			fixtures:    []string{fixtureActiveSessions},
			expected: Expected{
				activeSession: nil,
				err:           store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when active session is found",
			resolver:    store.SessionUIDResolver,
			value:       "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
			fixtures:    []string{fixtureActiveSessions},
			expected: Expected{
				activeSession: &models.ActiveSession{
					UID:      "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
					LastSeen: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
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

			activeSession, err := s.ActiveSessionResolve(ctx, tc.resolver, tc.value)
			assert.Equal(t, tc.expected, Expected{activeSession: activeSession, err: err})
		})
	}
}

func TestActiveSessionUpdate(t *testing.T) {
	cases := []struct {
		description   string
		activeSession *models.ActiveSession
		fixtures      []string
		expected      error
	}{
		{
			description: "fails when active session is not found",
			activeSession: &models.ActiveSession{
				UID:      "nonexistent",
				LastSeen: time.Now(),
			},
			fixtures: []string{fixtureActiveSessions},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when active session is found",
			activeSession: &models.ActiveSession{
				UID:      "a3b0431f5df6a7827945d2e34872a5c781452bc36de42f8b1297fd9ecb012f68",
				LastSeen: time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC),
			},
			fixtures: []string{fixtureActiveSessions},
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

			err := s.ActiveSessionUpdate(ctx, tc.activeSession)
			assert.Equal(t, tc.expected, err)
		})
	}
}
