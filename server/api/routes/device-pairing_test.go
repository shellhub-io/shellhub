package routes

import (
	"net/http"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestCreateDevicePairingBoundsFields(t *testing.T) {
	newRequest := func() *requests.DevicePairingCreate {
		return &requests.DevicePairingCreate{
			Info: &requests.DeviceInfo{
				ID:         "docker",
				PrettyName: "registry.example.com/team/app:1.0",
				Version:    "v0.31.1",
				Arch:       "amd64",
				Platform:   "connector",
			},
			Identity:  &requests.DeviceIdentity{MAC: "00:11:22:33:44:55"},
			PublicKey: "public-key",
		}
	}

	fields := func(req *requests.DevicePairingCreate) deviceFields {
		return deviceFields{info: req.Info, identity: req.Identity, publicKey: &req.PublicKey}
	}

	expectPairing := func(mock *mocks.MockService) {
		mock.On("CreateDevicePairing", gomock.Anything, gomock.AnythingOfType("*requests.DevicePairingCreate")).
			Return(&models.DevicePairing{}, nil).Once()
	}

	t.Run("accepts a connector agent with a 4096-bit key and an image reference at the limit", func(t *testing.T) {
		mock := mocks.NewMockService(t)
		expectPairing(mock)

		req := newRequest()
		req.Info.PrettyName = atLimitConnectorImage(t)
		req.PublicKey = rsa4096AgentPublicKey(t)

		assert.Equal(t, http.StatusOK, postDeviceJSON(t, mock, "/api/devices/pairing", req))
	})

	for _, bound := range deviceFieldBounds {
		t.Run(bound.field+" at its limit is accepted", func(t *testing.T) {
			mock := mocks.NewMockService(t)
			expectPairing(mock)

			req := newRequest()
			bound.set(fields(req), strings.Repeat("a", bound.limit))

			assert.Equal(t, http.StatusOK, postDeviceJSON(t, mock, "/api/devices/pairing", req))
		})

		t.Run(bound.field+" over its limit is refused before the service", func(t *testing.T) {
			mock := mocks.NewMockService(t)

			req := newRequest()
			bound.set(fields(req), strings.Repeat("a", bound.limit+1))

			assert.Equal(t, http.StatusBadRequest, postDeviceJSON(t, mock, "/api/devices/pairing", req))
		})
	}
}
