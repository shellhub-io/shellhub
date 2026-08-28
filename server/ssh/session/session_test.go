package session

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// stubContext is a minimal gliderssh.Context backed by a map for SetValue/Value
// (the snapshot helpers used by Evaluate rely on those round-tripping).
type stubContext struct {
	context.Context
	sync.Mutex
	values map[any]any
}

func newStubContext() *stubContext {
	return &stubContext{
		Context: context.Background(),
		values:  make(map[any]any),
	}
}

func (s *stubContext) User() string          { return "user@namespace.device" }
func (s *stubContext) SessionID() string     { return "test-session-id" }
func (s *stubContext) ClientVersion() string { return "" }
func (s *stubContext) ServerVersion() string { return "" }
func (s *stubContext) RemoteAddr() net.Addr  { return nil }
func (s *stubContext) LocalAddr() net.Addr   { return nil }
func (s *stubContext) Permissions() *gliderssh.Permissions {
	return &gliderssh.Permissions{}
}

func (s *stubContext) SetValue(key, val any) {
	s.Lock()
	defer s.Unlock()

	s.values[key] = val
}

func (s *stubContext) Value(key any) any {
	s.Lock()
	defer s.Unlock()

	return s.values[key]
}

// newTestSession builds a Session with only the fields Evaluate reads.
func newTestSession(service services.Service) *Session {
	tgt, _ := target.NewTarget("user@namespace.device") //nolint:errcheck

	return &Session{
		UID:     "test-uid",
		service: service,
		Data: Data{
			Target:    tgt,
			IPAddress: "127.0.0.1",
			SSHID:     "user@namespace.device",
			Device: &models.Device{
				UID:      "device-uid",
				Name:     "device",
				TenantID: "tenant-id",
			},
			Namespace: &models.Namespace{
				Name: "namespace",
			},
		},
		once:  new(sync.Once),
		seats: newSeats(),
		agent: &Agent{channels: make(map[int]*AgentChannel)},
		client: &Client{
			channels: make(map[int]*ClientChannel),
		},
	}
}

func TestRecorded(t *testing.T) {
	errStoreDown := errors.New("store is down")

	tests := []struct {
		description string
		record      bool
		pty         bool
		setupMock   func(m *servicemocks.MockService)
		expectedErr error
		// expectSkipped is whether the error means the session was never going to be
		// recorded, which is what the caller reports below a warning.
		expectSkipped bool
	}{
		{
			description:   "recording disabled for the namespace",
			record:        false,
			pty:           true,
			setupMock:     func(_ *servicemocks.MockService) {},
			expectedErr:   ErrRecordingDisabled,
			expectSkipped: true,
		},
		{
			description:   "seat has no pty to record",
			record:        true,
			pty:           false,
			setupMock:     func(_ *servicemocks.MockService) {},
			expectedErr:   ErrRecordingNoPty,
			expectSkipped: true,
		},
		{
			description: "marking the session as recorded fails",
			record:      true,
			pty:         true,
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					UpdateSession(mock.Anything, models.UID("test-uid"), mock.Anything).
					Return(errStoreDown).
					Once()
			},
			expectedErr:   errStoreDown,
			expectSkipped: false,
		},
		{
			description: "session is marked as recorded",
			record:      true,
			pty:         true,
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					UpdateSession(mock.Anything, models.UID("test-uid"), mock.Anything).
					Return(nil).
					Once()
			},
			expectedErr:   nil,
			expectSkipped: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			serviceMock := servicemocks.NewMockService(t)
			tt.setupMock(serviceMock)

			sess := newTestSession(serviceMock)
			sess.Namespace.Settings = &models.NamespaceSettings{SessionRecord: tt.record}

			seat, err := sess.seats.NewSeat()
			require.NoError(t, err)
			sess.seats.SetPty(seat, tt.pty)

			err = sess.Recorded(seat)

			require.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expectSkipped, errors.Is(err, ErrRecordingSkipped))
		})
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		description          string
		edition              envs.Edition
		setupMock            func(m *servicemocks.MockService)
		expectedErr          error
		expectStateEvaluated bool
	}{
		{
			description: "cloud: firewall denies the connection",
			edition:     envs.Cloud,
			setupMock: func(m *servicemocks.MockService) {
				// Firewall runs first; a denial stops before billing is consulted.
				m.EXPECT().
					EvaluateFirewall(mock.Anything, mock.Anything).
					Return(services.ErrFirewallBlocked).
					Once()
			},
			expectedErr:          ErrFirewallBlock,
			expectStateEvaluated: false,
		},
		{
			description: "cloud: a firewall evaluation failure is not a rule block",
			edition:     envs.Cloud,
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					EvaluateFirewall(mock.Anything, mock.Anything).
					Return(errors.New("no license to evaluate the firewall against")).
					Once()
			},
			expectedErr:          ErrFirewallUnknown,
			expectStateEvaluated: false,
		},
		{
			description: "cloud: firewall allows and billing succeeds",
			edition:     envs.Cloud,
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					EvaluateFirewall(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				// The tenant comes from the device the session already resolved, so
				// billing is evaluated without a second device lookup.
				m.EXPECT().
					EvaluateBilling(mock.Anything, "tenant-id").
					Return(nil).
					Once()
			},
			expectedErr:          nil,
			expectStateEvaluated: true,
		},
		{
			description: "cloud: firewall allows but billing blocks",
			edition:     envs.Cloud,
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					EvaluateFirewall(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				m.EXPECT().
					EvaluateBilling(mock.Anything, "tenant-id").
					Return(services.ErrBillingBlocked).
					Once()
			},
			expectedErr:          ErrBillingBlock,
			expectStateEvaluated: false,
		},
		{
			description: "enterprise (not cloud): license blocks before the firewall runs",
			edition:     envs.Enterprise,
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					EvaluateLicense(mock.Anything).
					Return(services.ErrLicenseBlocked).
					Once()
			},
			expectedErr:          ErrLicenseBlock,
			expectStateEvaluated: false,
		},
		{
			description: "enterprise (not cloud): license allows, firewall denies",
			edition:     envs.Enterprise,
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					EvaluateLicense(mock.Anything).
					Return(nil).
					Once()

				m.EXPECT().
					EvaluateFirewall(mock.Anything, mock.Anything).
					Return(services.ErrFirewallBlocked).
					Once()
			},
			expectedErr:          ErrFirewallBlock,
			expectStateEvaluated: false,
		},
		{
			description: "enterprise (not cloud): license allows, firewall allows",
			edition:     envs.Enterprise,
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					EvaluateLicense(mock.Anything).
					Return(nil).
					Once()

				m.EXPECT().
					EvaluateFirewall(mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			expectedErr:          nil,
			expectStateEvaluated: true,
		},
		{
			description: "community: no firewall, billing, or license evaluation",
			edition:     envs.Community,
			setupMock: func(_ *servicemocks.MockService) {
				// no gate is consulted in community mode.
			},
			expectedErr:          nil,
			expectStateEvaluated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			envstest.SetEdition(t, tt.edition)

			serviceMock := servicemocks.NewMockService(t)
			tt.setupMock(serviceMock)

			sess := newTestSession(serviceMock)
			ctx := newStubContext()

			snap := getSnapshot(ctx)
			snap.save(sess, StateCreated)

			err := sess.Evaluate(ctx)

			require.ErrorIs(t, err, tt.expectedErr)

			_, state := snap.retrieve()
			if tt.expectStateEvaluated {
				assert.Equal(t, StateEvaluated, state, "expected snapshot to advance to StateEvaluated on success")
			} else {
				assert.Equal(t, StateCreated, state, "snapshot must remain StateCreated when Evaluate returns an error")
			}
		})
	}
}
