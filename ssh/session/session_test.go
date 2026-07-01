package session

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mapBackend is a test-only envs.Backend backed by a plain map.
type mapBackend struct {
	values map[string]string
}

func (m *mapBackend) Get(key string) string { return m.values[key] }

func (m *mapBackend) Process(_ string, _ interface{}) error { return nil }

// boolToEnv converts a bool to the string value envs checks (ENABLED or "").
func boolToEnv(v bool) string {
	if v {
		return envs.ENABLED
	}

	return ""
}

// stubContext is a minimal gliderssh.Context backed by a map for SetValue/Value
// (the snapshot helpers used by Evaluate rely on those round-tripping).
type stubContext struct {
	context.Context
	sync.Mutex
	values map[interface{}]interface{}
}

func newStubContext() *stubContext {
	return &stubContext{
		Context: context.Background(),
		values:  make(map[interface{}]interface{}),
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

func (s *stubContext) SetValue(key, val interface{}) {
	s.Lock()
	defer s.Unlock()

	s.values[key] = val
}

func (s *stubContext) Value(key interface{}) interface{} {
	s.Lock()
	defer s.Unlock()

	return s.values[key]
}

// newTestSession builds a Session with only the fields Evaluate reads.
func newTestSession(apiClient internalclient.Client) *Session {
	tgt, _ := target.NewTarget("user@namespace.device") //nolint:errcheck

	return &Session{
		UID: "test-uid",
		api: apiClient,
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
		Seats: NewSeats(),
		Agent: &Agent{Channels: make(map[int]*AgentChannel)},
		Client: &Client{
			Channels: make(map[int]*ClientChannel),
		},
	}
}

// TestEvaluate covers the firewall enforcement gate across deployment modes.
// The cloud cases are the regression for shellhub-io/team#166: before the fix,
// checkFirewall was never called when SHELLHUB_CLOUD=true, so firewall rules
// were silently ignored. The order of checks is firewall, then billing.
func TestEvaluate(t *testing.T) {
	type envConfig struct {
		isCloud      bool
		isEnterprise bool
	}

	tests := []struct {
		description          string
		env                  envConfig
		setupMock            func(m *mocks.MockClient)
		expectedErr          error
		expectStateEvaluated bool
	}{
		// ---- cloud (SHELLHUB_ENTERPRISE=true, SHELLHUB_CLOUD=true) ----
		{
			description: "cloud: firewall denies the connection",
			env:         envConfig{isCloud: true, isEnterprise: true},
			setupMock: func(m *mocks.MockClient) {
				// Firewall runs first; a 403 denies before billing is consulted.
				m.EXPECT().
					FirewallEvaluate(mock.Anything, mock.Anything).
					Return(&internalclient.Error{Code: http.StatusForbidden}).
					Once()
			},
			expectedErr:          ErrFirewallBlock,
			expectStateEvaluated: false,
		},
		{
			description: "cloud: firewall allows and billing succeeds",
			env:         envConfig{isCloud: true, isEnterprise: true},
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().
					FirewallEvaluate(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				m.EXPECT().
					GetDevice(mock.Anything, "device-uid").
					Return(&models.Device{UID: "device-uid", TenantID: "billing-tenant-id"}, nil).
					Once()

				m.EXPECT().
					BillingEvaluate(mock.Anything, "billing-tenant-id").
					Return(&models.BillingEvaluation{CanConnect: true}, nil).
					Once()
			},
			expectedErr:          nil,
			expectStateEvaluated: true,
		},
		{
			description: "cloud: firewall allows but billing blocks",
			env:         envConfig{isCloud: true, isEnterprise: true},
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().
					FirewallEvaluate(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				m.EXPECT().
					GetDevice(mock.Anything, "device-uid").
					Return(&models.Device{UID: "device-uid", TenantID: "billing-tenant-id"}, nil).
					Once()

				m.EXPECT().
					BillingEvaluate(mock.Anything, "billing-tenant-id").
					Return(&models.BillingEvaluation{CanConnect: false}, nil).
					Once()
			},
			expectedErr:          ErrBillingBlock,
			expectStateEvaluated: false,
		},
		// ---- enterprise-only (SHELLHUB_ENTERPRISE=true, SHELLHUB_CLOUD=false) ----
		{
			description: "enterprise (not cloud): license allows, firewall denies",
			env:         envConfig{isCloud: false, isEnterprise: true},
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().
					LicenseEvaluate(mock.Anything).
					Return(&models.BillingEvaluation{CanConnect: true}, nil).
					Once()

				m.EXPECT().
					FirewallEvaluate(mock.Anything, mock.Anything).
					Return(&internalclient.Error{Code: http.StatusForbidden}).
					Once()
			},
			expectedErr:          ErrFirewallBlock,
			expectStateEvaluated: false,
		},
		{
			description: "enterprise (not cloud): license allows, firewall allows",
			env:         envConfig{isCloud: false, isEnterprise: true},
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().
					LicenseEvaluate(mock.Anything).
					Return(&models.BillingEvaluation{CanConnect: true}, nil).
					Once()

				m.EXPECT().
					FirewallEvaluate(mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			expectedErr:          nil,
			expectStateEvaluated: true,
		},
		// ---- community (both flags false) ----
		{
			description: "community: no firewall, billing, or license evaluation",
			env:         envConfig{isCloud: false, isEnterprise: false},
			setupMock: func(_ *mocks.MockClient) {
				// no API calls expected in community mode.
			},
			expectedErr:          nil,
			expectStateEvaluated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			prev := envs.DefaultBackend
			t.Cleanup(func() { envs.DefaultBackend = prev })

			envs.DefaultBackend = &mapBackend{
				values: map[string]string{
					"SHELLHUB_ENTERPRISE": boolToEnv(tt.env.isEnterprise),
					"SHELLHUB_CLOUD":      boolToEnv(tt.env.isCloud),
				},
			}

			apiMock := mocks.NewMockClient(t)
			tt.setupMock(apiMock)

			sess := newTestSession(apiMock)
			ctx := newStubContext()

			snap := getSnapshot(ctx)
			snap.save(sess, StateCreated)

			err := sess.Evaluate(ctx)

			assert.ErrorIs(t, err, tt.expectedErr)

			_, state := snap.retrieve()
			if tt.expectStateEvaluated {
				assert.EqualValues(t, StateEvaluated, state, "expected snapshot to advance to StateEvaluated on success")
			} else {
				assert.EqualValues(t, StateCreated, state, "snapshot must remain StateCreated when Evaluate returns an error")
			}
		})
	}
}
