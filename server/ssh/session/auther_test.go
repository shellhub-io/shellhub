package session

import (
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newAgentVersionSession(service *servicemocks.MockService, version string) *Session {
	sess := newTestSession(service)
	sess.Device.Info = &models.DeviceInfo{Version: version}

	return sess
}

func withAllowPublickeyAccessBelow060(t *testing.T, allow bool) {
	t.Helper()

	previous := *sshconf
	t.Cleanup(func() { Configure(previous) })

	next := previous
	next.AllowPublickeyAccessBelow060 = allow
	Configure(next)
}

func TestPublicKeyOfferAgentVersionFloor(t *testing.T) {
	cases := []struct {
		name        string
		version     string
		allowBelow  bool
		expectedErr error
	}{
		{
			name:        "refuses an agent older than 0.6.0",
			version:     "0.5.9",
			expectedErr: ErrUnsuportedPublicKeyAuth,
		},
		{
			name:        "refuses an agent older than 0.6.0 reported with a v prefix",
			version:     "v0.5.9",
			expectedErr: ErrUnsuportedPublicKeyAuth,
		},
		{
			name:        "refuses the oldest agent that ever reported a version",
			version:     "0.0.1",
			expectedErr: ErrUnsuportedPublicKeyAuth,
		},
		{
			name:        "refuses a version it cannot parse",
			version:     "not-a-version",
			expectedErr: ErrInvalidVersion,
		},
		{
			name:        "refuses an agent reporting no version at all",
			version:     "",
			expectedErr: ErrInvalidVersion,
		},
		{
			name:       "allows an agent older than 0.6.0 when the instance opts in",
			version:    "0.5.9",
			allowBelow: true,
		},
		{
			name:       "allows a version it cannot parse when the instance opts in",
			version:    "not-a-version",
			allowBelow: true,
		},
		{
			name:    "allows exactly 0.6.0",
			version: "0.6.0",
		},
		{
			name:    "allows an agent newer than the floor",
			version: "0.19.4",
		},
		{
			name:    "allows an agent reporting latest",
			version: "latest",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withAllowPublickeyAccessBelow060(t, tc.allowBelow)

			serviceMock := servicemocks.NewMockService(t)
			if tc.expectedErr == nil {
				expectKeyAdmitted(serviceMock)
			}

			sess := newAgentVersionSession(serviceMock, tc.version)

			err := AuthPublicKey(newTestSSHKey(t)).Offer(sess)

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPublicKeyOfferKeyEvaluation(t *testing.T) {
	errStore := errors.New("store is unreachable")
	key := &models.PublicKey{Fingerprint: "fingerprint"}

	cases := []struct {
		name        string
		mocks       func(*servicemocks.MockService)
		expectedErr error
	}{
		{
			name:  "admits a key the username and the filter both accept",
			mocks: expectKeyAdmitted,
		},
		{
			name: "refuses a key the username rejects",
			mocks: func(s *servicemocks.MockService) {
				s.EXPECT().GetPublicKey(mock.Anything, mock.Anything, "tenant-id").Return(key, nil).Once()
				s.EXPECT().EvaluateKeyUsername(mock.Anything, key, "user").Return(false, nil).Once()
				s.EXPECT().EvaluateKeyFilter(mock.Anything, key, mock.Anything).Return(true, nil).Once()
			},
			expectedErr: ErrEvaluatePublicKey,
		},
		{
			name: "refuses a key the filter rejects",
			mocks: func(s *servicemocks.MockService) {
				s.EXPECT().GetPublicKey(mock.Anything, mock.Anything, "tenant-id").Return(key, nil).Once()
				s.EXPECT().EvaluateKeyUsername(mock.Anything, key, "user").Return(true, nil).Once()
				s.EXPECT().EvaluateKeyFilter(mock.Anything, key, mock.Anything).Return(false, nil).Once()
			},
			expectedErr: ErrEvaluatePublicKey,
		},
		{
			name: "refuses a key whose username evaluation fails",
			mocks: func(s *servicemocks.MockService) {
				s.EXPECT().GetPublicKey(mock.Anything, mock.Anything, "tenant-id").Return(key, nil).Once()
				s.EXPECT().EvaluateKeyUsername(mock.Anything, key, "user").Return(false, errStore).Once()
			},
			expectedErr: ErrEvaluatePublicKey,
		},
		{
			name: "refuses a key whose filter evaluation fails",
			mocks: func(s *servicemocks.MockService) {
				s.EXPECT().GetPublicKey(mock.Anything, mock.Anything, "tenant-id").Return(key, nil).Once()
				s.EXPECT().EvaluateKeyUsername(mock.Anything, key, "user").Return(true, nil).Once()
				s.EXPECT().EvaluateKeyFilter(mock.Anything, key, mock.Anything).Return(false, errStore).Once()
			},
			expectedErr: ErrEvaluatePublicKey,
		},
		{
			name: "surfaces the lookup error for a key the namespace does not hold",
			mocks: func(s *servicemocks.MockService) {
				s.EXPECT().GetPublicKey(mock.Anything, mock.Anything, "tenant-id").Return(nil, errStore).Once()
			},
			expectedErr: errStore,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withAllowPublickeyAccessBelow060(t, false)

			serviceMock := servicemocks.NewMockService(t)
			tc.mocks(serviceMock)

			sess := newAgentVersionSession(serviceMock, "latest")

			err := AuthPublicKey(newTestSSHKey(t)).Offer(sess)

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func expectKeyAdmitted(service *servicemocks.MockService) {
	key := &models.PublicKey{Fingerprint: "fingerprint"}

	service.EXPECT().GetPublicKey(mock.Anything, mock.Anything, "tenant-id").Return(key, nil).Once()
	service.EXPECT().EvaluateKeyUsername(mock.Anything, key, "user").Return(true, nil).Once()
	service.EXPECT().EvaluateKeyFilter(mock.Anything, key, mock.Anything).Return(true, nil).Once()
}
