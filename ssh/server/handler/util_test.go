package handler

import (
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	metadataMocks "github.com/shellhub-io/shellhub/ssh/pkg/metadata/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCheckAgentVersionForPublicKey(t *testing.T) {
	cases := []struct {
		description   string
		requiredMocks func(gliderssh.Context)
		expected      error
	}{
		{
			description: "succeeds when authentication method is passwod",
			requiredMocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("RestoreDevice", ctx).
					Return(&models.Device{
						Info: &models.DeviceInfo{
							Version: "latest",
						},
					}).
					Once()
			},
			expected: nil,
		},
		{
			description: "succeeds when device's version is 'latest'",
			requiredMocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("RestoreDevice", ctx).
					Return(&models.Device{
						Info: &models.DeviceInfo{
							Version: "latest",
						},
					}).
					Once()
			},
			expected: nil,
		},
		{
			description: "fails when device's version is 0.5.x or earlier and authentication method is PUBLIC KEY",
			requiredMocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.SetBackend(metadataMock)

				metadataMock.On("RestoreDevice", ctx).
					Return(&models.Device{
						Info: &models.DeviceInfo{
							Version: "0.5.2",
						},
					}).
					Once()
			},
			expected: ErrUnsuportedPublicKeyAuth,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := new(gliderssh.Context)
			tc.requiredMocks(*ctx)

			err := checkAgentVersionForPublicKey(*ctx)
			assert.Equal(t, tc.expected, err)
		})
	}
}
