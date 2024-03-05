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
	"github.com/shellhub-io/shellhub/ssh/pkg/sshsrvtest"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gossh "golang.org/x/crypto/ssh"
)

func generateTestPubKey(t *testing.T) gliderssh.PublicKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	publicRsaKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	return publicRsaKey
}

func TestPublicKeyHandler(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		mocks       func(ctx gliderssh.Context, publicKey gliderssh.PublicKey)
		expected    bool
	}{
		{
			description: "fails when could not store the target",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				srv := sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
						},
					},
					&gossh.ClientConfig{
						User:            "user@namespace.00-00-00-00-00-00",
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)

				return srv
			},
			mocks: func(ctx gliderssh.Context, _ gliderssh.PublicKey) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
					Once()

				metadataMock.On("MaybeStoreTarget", ctx, "user@namespace.00-00-00-00-00-00").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: false,
		},
		{
			description: "fails when could not store the lookup",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				srv := sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
						},
					},
					&gossh.ClientConfig{
						User:            "user@namespace.00-00-00-00-00-00",
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)

				return srv
			},
			mocks: func(ctx gliderssh.Context, _ gliderssh.PublicKey) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
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
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				srv := sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
						},
					},
					&gossh.ClientConfig{
						User:            "user@namespace.00-00-00-00-00-00",
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)

				return srv
			},
			mocks: func(ctx gliderssh.Context, _ gliderssh.PublicKey) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
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
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				srv := sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
						},
					},
					&gossh.ClientConfig{
						User:            "user@namespace.00-00-00-00-00-00",
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)

				return srv
			},
			mocks: func(ctx gliderssh.Context, publicKey gliderssh.PublicKey) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
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

				fingerprint := gossh.FingerprintLegacyMD5(publicKey)

				api.On("GetPublicKey", fingerprint, "00000000-0000-4000-0000-000000000000").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: false,
		},
		{
			description: "fails when could not evaluate the key",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				srv := sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
						},
					},
					&gossh.ClientConfig{
						User:            "user@namespace.00-00-00-00-00-00",
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)

				return srv
			},
			mocks: func(ctx gliderssh.Context, publicKey gliderssh.PublicKey) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
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

				fingerprint := gossh.FingerprintLegacyMD5(publicKey)

				api.On("GetPublicKey", fingerprint, "00000000-0000-4000-0000-000000000000").
					Return(nil, nil).
					Once()

				api.On("EvaluateKey", fingerprint, &models.Device{TenantID: "00000000-0000-4000-0000-000000000000"}, "user").
					Return(false, errors.New("error")).
					Once()
			},
			expected: false,
		},
		{
			description: "succeeds to authenticate the session",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				srv := sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
						},
					},
					&gossh.ClientConfig{
						User:            "user@namespace.00-00-00-00-00-00",
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)

				return srv
			},
			mocks: func(ctx gliderssh.Context, publicKey gliderssh.PublicKey) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("MaybeStoreSSHID", ctx, "user@namespace.00-00-00-00-00-00").
					Return("user@namespace.00-00-00-00-00-00").
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

				fingerprint := gossh.FingerprintLegacyMD5(publicKey)

				api.On("GetPublicKey", fingerprint, "00000000-0000-4000-0000-000000000000").
					Return(nil, nil).
					Once()

				api.On("EvaluateKey", fingerprint, &models.Device{TenantID: "00000000-0000-4000-0000-000000000000"}, "user").
					Return(true, nil).
					Once()
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)
			defer srv.Teardown()

			srv.Start()
			assert.NoError(t, srv.Agent.Run("cmd"))

			publicKey := generateTestPubKey(t)

			tc.mocks(ctx, publicKey)

			result := publicKeyHandler(ctx, publicKey)
			assert.Equal(t, tc.expected, result)
		})
	}
}
