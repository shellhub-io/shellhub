package session

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	gossh "golang.org/x/crypto/ssh"
)

func TestPasswordAuthEvaluate(t *testing.T) {
	cases := []struct {
		name          string
		allowPassword bool
		expectedError error
	}{
		{
			name:          "password auth enabled",
			allowPassword: true,
			expectedError: nil,
		},
		{
			name:          "password auth disabled",
			allowPassword: false,
			expectedError: ErrPasswordDisabled,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess := &Session{
				Data: Data{
					Target: &target.Target{Username: "user"},
					Namespace: &models.Namespace{
						Settings: &models.NamespaceSettings{
							AllowPassword: tc.allowPassword,
						},
					},
				},
			}

			auth := AuthPassword("password")
			err := auth.Evaluate(sess)

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestPublicKeyAuthEvaluate(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	publicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	fingerprint := gossh.FingerprintLegacyMD5(publicKey)

	cases := []struct {
		name           string
		allowPublicKey bool
		mockSetup      func(*mocks.Client)
		expectedError  error
	}{
		{
			name:           "public key auth enabled",
			allowPublicKey: true,
			mockSetup: func(m *mocks.Client) {
				m.On("GetPublicKey", testifymock.Anything, fingerprint, "tenant-1").Return(nil, nil)
				m.On("EvaluateKey", testifymock.Anything, fingerprint, testifymock.Anything, testifymock.Anything).Return(true, nil)
			},
			expectedError: nil,
		},
		{
			name:           "public key auth disabled",
			allowPublicKey: false,
			mockSetup:      func(*mocks.Client) {},
			expectedError:  ErrPublicKeyDisabled,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mocks.Client{}
			tc.mockSetup(mockClient)

			sess := &Session{
				Data: Data{
					Target: &target.Target{Username: "user"},
					Device: &models.Device{
						Info:     &models.DeviceInfo{Version: "latest"},
						TenantID: "tenant-1",
					},
					Namespace: &models.Namespace{
						Settings: &models.NamespaceSettings{
							AllowPublicKey: tc.allowPublicKey,
						},
					},
				},
				api: mockClient,
			}

			auth := AuthPublicKey(publicKey)
			err := auth.Evaluate(sess)

			assert.Equal(t, tc.expectedError, err)
		})
	}
}
