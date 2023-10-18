package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	internalclientMocks "github.com/shellhub-io/shellhub/pkg/api/internalclient/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	metadataMocks "github.com/shellhub-io/shellhub/ssh/pkg/metadata/mocks"
	"github.com/shellhub-io/shellhub/ssh/pkg/sshtest"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gossh "golang.org/x/crypto/ssh"
)

// generateTestKey generates a new test SSH public key.
func generateTestKey() (gliderssh.PublicKey, error) {
	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Extract the public key and convert it into ssh.PublicKey format.
	publicRsaKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return publicRsaKey, nil
}

func TestPublicKeyHandler(t *testing.T) {
	cases := []struct {
		description string
		mocks       func(ctx gliderssh.Context)
		expected    bool
	}{
		{
			description: "fails when could not store the target",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
					Once()

				metadataMock.On("MaybeStoreFingerprint", ctx, mock.Anything).
					Return("fingerprint").
					Once()

				metadataMock.On("MaybeStoreTarget", ctx, "user@namespace.00-00-00-00-00-00").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: false,
		},
		{
			description: "fails when could not store the lookup",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
					Once()

				metadataMock.On("MaybeStoreFingerprint", ctx, mock.Anything).
					Return("fingerprint").
					Once()

				tag := &target.Target{Username: "user", Data: "namespace.00-00-00-00-00-00"}
				metadataMock.On("MaybeStoreTarget", ctx, "user@namespace.00-00-00-00-00-00").
					Return(tag, nil).
					Once()

				api := new(internalclientMocks.Client)
				// Since MaybeSetAPI uses `internalclient.NewClient()` as an argument, using `api` here would result in a memory error.
				metadataMock.On("MaybeSetAPI", ctx, mock.Anything).
					Return(api).
					Once()

				// lookup := map[string]string{}
				metadataMock.On("MaybeStoreLookup", ctx, tag, api).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: false,
		},
		{
			description: "fails when could not store the device",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
					Once()

				metadataMock.On("MaybeStoreFingerprint", ctx, mock.Anything).
					Return("fingerprint").
					Once()

				tag := &target.Target{Username: "user", Data: "namespace.00-00-00-00-00-00"}
				metadataMock.On("MaybeStoreTarget", ctx, "user@namespace.00-00-00-00-00-00").
					Return(tag, nil).
					Once()

				api := new(internalclientMocks.Client)
				// Since MaybeSetAPI uses `internalclient.NewClient()` as an argument, using `api` here would result in a memory error.
				metadataMock.On("MaybeSetAPI", ctx, mock.Anything).
					Return(api).
					Once()

				lookup := map[string]string{}
				metadataMock.On("MaybeStoreLookup", ctx, tag, api).
					Return(lookup, nil).
					Once()

				metadataMock.On("MaybeStoreDevice", ctx, lookup, api).
					Return(nil, []error{errors.New("error")}).
					Once()
			},
			expected: false,
		},
		{
			description: "fails when could not get the public key",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
					Once()

				metadataMock.On("MaybeStoreFingerprint", ctx, mock.Anything).
					Return("fingerprint").
					Once()

				tag := &target.Target{Username: "user", Data: "namespace.00-00-00-00-00-00"}
				metadataMock.On("MaybeStoreTarget", ctx, "user@namespace.00-00-00-00-00-00").
					Return(tag, nil).
					Once()

				api := new(internalclientMocks.Client)
				// Since MaybeSetAPI uses `internalclient.NewClient()` as an argument, using `api` here would result in a memory error.
				metadataMock.On("MaybeSetAPI", ctx, mock.Anything).
					Return(api).
					Once()

				lookup := map[string]string{}
				metadataMock.On("MaybeStoreLookup", ctx, tag, api).
					Return(lookup, nil).
					Once()

				metadataMock.On("MaybeStoreDevice", ctx, lookup, api).
					Return(&models.Device{TenantID: "00000000-0000-4000-0000-000000000000"}, []error{}).
					Once()

				api.On("GetPublicKey", "fingerprint", "00000000-0000-4000-0000-000000000000").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: false,
		},
		{
			description: "fails when could not evaluate the key",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
					Once()

				metadataMock.On("MaybeStoreFingerprint", ctx, mock.Anything).
					Return("fingerprint").
					Once()

				tag := &target.Target{Username: "user", Data: "namespace.00-00-00-00-00-00"}
				metadataMock.On("MaybeStoreTarget", ctx, "user@namespace.00-00-00-00-00-00").
					Return(tag, nil).
					Once()

				api := new(internalclientMocks.Client)
				// Since MaybeSetAPI uses `internalclient.NewClient()` as an argument, using `api` here would result in a memory error.
				metadataMock.On("MaybeSetAPI", ctx, mock.Anything).
					Return(api).
					Once()

				lookup := map[string]string{}
				metadataMock.On("MaybeStoreLookup", ctx, tag, api).
					Return(lookup, nil).
					Once()

				metadataMock.On("MaybeStoreDevice", ctx, lookup, api).
					Return(&models.Device{TenantID: "00000000-0000-4000-0000-000000000000"}, []error{}).
					Once()

				api.On("GetPublicKey", "fingerprint", "00000000-0000-4000-0000-000000000000").
					Return(nil, nil).
					Once()

				api.On("EvaluateKey", "fingerprint", &models.Device{TenantID: "00000000-0000-4000-0000-000000000000"}, "user").
					Return(false, errors.New("error")).
					Once()
			},
			expected: false,
		},
		{
			description: "succeeds to authenticate the session",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
					Once()

				metadataMock.On("MaybeStoreFingerprint", ctx, mock.Anything).
					Return("fingerprint").
					Once()

				tag := &target.Target{Username: "user", Data: "namespace.00-00-00-00-00-00"}
				metadataMock.On("MaybeStoreTarget", ctx, "user@namespace.00-00-00-00-00-00").
					Return(tag, nil).
					Once()

				api := new(internalclientMocks.Client)
				// Since MaybeSetAPI uses `internalclient.NewClient()` as an argument, using `api` here would result in a memory error.
				metadataMock.On("MaybeSetAPI", ctx, mock.Anything).
					Return(api).
					Once()

				lookup := map[string]string{}
				metadataMock.On("MaybeStoreLookup", ctx, tag, api).
					Return(lookup, nil).
					Once()

				metadataMock.On("MaybeStoreDevice", ctx, lookup, api).
					Return(&models.Device{TenantID: "00000000-0000-4000-0000-000000000000"}, []error{}).
					Once()

				api.On("GetPublicKey", "fingerprint", "00000000-0000-4000-0000-000000000000").
					Return(nil, nil).
					Once()

				api.On("EvaluateKey", "fingerprint", &models.Device{TenantID: "00000000-0000-4000-0000-000000000000"}, "user").
					Return(true, nil).
					Once()

				metadataMock.On("StoreAuthenticationMethod", ctx, metadata.AuthMethodPubKey)
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context
			sshConn := sshtest.Start(
				t,
				&gossh.ClientConfig{
					User: "user@namespace.00-00-00-00-00-00",
					Auth: []gossh.AuthMethod{
						gossh.Password("secret"),
					},
				},
				&gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						ctx = s.Context()
					},
				},
			)
			defer sshConn.Teardown()

			assert.NoError(t, sshConn.Session.Run("cmd"))
			tc.mocks(ctx)

			publicKey, err := generateTestKey()
			assert.NoError(t, err)

			result := PublicKeyHandler(ctx, publicKey)
			assert.Equal(t, tc.expected, result)
		})
	}
}
