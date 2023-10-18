package auth

import (
	"errors"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	internalclientMocks "github.com/shellhub-io/shellhub/pkg/api/internalclient/mocks"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	metadataMocks "github.com/shellhub-io/shellhub/ssh/pkg/metadata/mocks"
	"github.com/shellhub-io/shellhub/ssh/pkg/sshtest"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gossh "golang.org/x/crypto/ssh"
)

func TestPasswordHandler(t *testing.T) {
	cases := []struct {
		description string
		password    string
		mocks       func(ctx gliderssh.Context)
		expected    bool
	}{
		{
			description: "fails when could not store the target",
			password:    "secret",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

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
			password:    "secret",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

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
			password:    "secret",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

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
			description: "succeeds to authenticate the session",
			password:    "secret",
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

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
					Return(nil, []error{}).
					Once()

				metadataMock.On("StorePassword", ctx, "secret")
				metadataMock.On("StoreAuthenticationMethod", ctx, metadata.AuthMethodPasswd)
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

			result := PasswordHandler(ctx, tc.password)
			assert.Equal(t, tc.expected, result)
		})
	}
}
