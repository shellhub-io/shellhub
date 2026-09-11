package client

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	resty "github.com/go-resty/resty/v2"
	mock "github.com/jarcoal/httpmock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func authRequest() *models.DeviceAuthRequest {
	return &models.DeviceAuthRequest{
		Info: &models.DeviceInfo{ID: "manjaro", PrettyName: "Manjaro", Version: "latest", Arch: "amd64"},
		DeviceAuth: &models.DeviceAuth{
			Hostname: "83-18-77-25-78-0d",
			Identity: &models.DeviceIdentity{MAC: "83:18:77:25:78:0d"},
			TenantID: "00000000-0000-4000-0000-000000000000",
		},
	}
}

func TestRetriedAttemptsAreReportedAsWhatTheServerDid(t *testing.T) {
	tests := []struct {
		description string
		status      int
		expected    string
	}{
		{
			description: "a server too busy to answer is unreachable, not a refusal of the device",
			status:      http.StatusTooManyRequests,
			expected:    "Cannot reach the server, retrying until it answers",
		},
		{
			description: "a server that failed on its own side is unreachable, not a refusal of the device",
			status:      http.StatusInternalServerError,
			expected:    "Cannot reach the server, retrying until it answers",
		},
		{
			description: "a server that is up and rejecting the device is a refusal",
			status:      http.StatusNotFound,
			expected:    "Cannot authorize the device, retrying until the server accepts it",
		},
		{
			description: "a namespace at its device limit is a refusal",
			status:      http.StatusPaymentRequired,
			expected:    "Cannot authorize the device, retrying until the server accepts it",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			backend, hook := logtest.NewNullLogger()
			backend.SetLevel(logrus.DebugLevel)

			cli, err := NewClient("https://www.cloud.shellhub.io/", withImmediateRetries(), WithLogger(backend))
			require.NoError(t, err)

			client, ok := cli.(*client)
			require.True(t, ok)

			mock.ActivateNonDefault(client.http.GetClient())
			defer mock.DeactivateAndReset()

			accepted, _ := mock.NewJsonResponder(200, models.DeviceAuthResponse{Name: "83-18-77-25-78-0d"})
			mock.RegisterResponder("POST", "/api/devices/auth",
				mock.NewStringResponder(test.status, `{"message":"nope"}`).Then(accepted))

			_, err = cli.AuthDevice(authRequest())
			require.NoError(t, err)

			require.NotEmpty(t, hook.AllEntries())
			assert.Equal(t, test.expected, hook.AllEntries()[0].Message)
			assert.Equal(t, logrus.WarnLevel, hook.AllEntries()[0].Level)
		})
	}
}

func TestForbiddenIsRetriedAsSomethingAnOperatorClears(t *testing.T) {
	cli, err := NewClient("https://www.cloud.shellhub.io/", withImmediateRetries())
	require.NoError(t, err)

	client, ok := cli.(*client)
	require.True(t, ok)

	mock.ActivateNonDefault(client.http.GetClient())
	defer mock.DeactivateAndReset()

	accepted, _ := mock.NewJsonResponder(200, models.DeviceAuthResponse{Name: "83-18-77-25-78-0d"})
	mock.RegisterResponder("POST", "/api/devices/auth",
		mock.NewStringResponder(http.StatusForbidden, `{"message":"device limit reached"}`).Then(accepted))

	response, err := cli.AuthDevice(authRequest())
	require.NoError(t, err)
	assert.NotNil(t, response)

	calls := 0
	for _, count := range mock.GetCallCountInfo() {
		calls += count
	}

	assert.Equal(t, 2, calls)
}

func TestARequestThatNeverLeftIsNotRetried(t *testing.T) {
	cli, err := NewClient("https://www.cloud.shellhub.io/", withImmediateRetries())
	require.NoError(t, err)

	client, ok := cli.(*client)
	require.True(t, ok)

	client.http.SetRetryCount(3)

	rejected := errors.New("the request never left")

	runs := 0
	client.http.OnBeforeRequest(func(_ *resty.Client, _ *resty.Request) error {
		runs++

		return rejected
	})

	mock.ActivateNonDefault(client.http.GetClient())
	defer mock.DeactivateAndReset()

	_, err = cli.GetInfo("v0.13.0")
	require.ErrorIs(t, err, rejected)

	assert.Equal(t, 1, runs)
}

func TestARefusalDoesNotReportThatTheServerCameBack(t *testing.T) {
	backend, hook := logtest.NewNullLogger()
	backend.SetLevel(logrus.DebugLevel)

	cli, err := NewClient("https://www.cloud.shellhub.io/", withImmediateRetries(), WithLogger(backend))
	require.NoError(t, err)

	client, ok := cli.(*client)
	require.True(t, ok)

	mock.ActivateNonDefault(client.http.GetClient())
	defer mock.DeactivateAndReset()

	mock.RegisterResponder("POST", "/api/devices/auth",
		mock.NewStringResponder(http.StatusNotFound, `{"message":"no such namespace"}`).
			Then(mock.NewStringResponder(http.StatusBadRequest, `{"message":"malformed"}`)))

	_, err = cli.AuthDevice(authRequest())
	require.Error(t, err)

	for _, entry := range hook.AllEntries() {
		assert.NotEqual(t, "Recovered after retrying", entry.Message)
	}
}

func TestARecoveryIsReportedOnlyAfterAnOutage(t *testing.T) {
	tests := []struct {
		description string
		outage      bool
	}{
		{
			description: "says nothing when the first attempt already answered",
			outage:      false,
		},
		{
			description: "reports the attempt that ended the outage",
			outage:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			backend, hook := logtest.NewNullLogger()
			backend.SetLevel(logrus.DebugLevel)

			cli, err := NewClient("https://www.cloud.shellhub.io/", withImmediateRetries(), WithLogger(backend))
			require.NoError(t, err)

			client, ok := cli.(*client)
			require.True(t, ok)

			mock.ActivateNonDefault(client.http.GetClient())
			defer mock.DeactivateAndReset()

			answered, _ := mock.NewJsonResponder(200, models.Info{Version: "v0.13.0"})

			responder := answered
			if test.outage {
				responder = mock.NewErrorResponder(errors.New("dial tcp: connection refused")).Then(answered)
			}

			mock.RegisterResponder("GET", "/info?agent_version=v0.13.0", responder)

			_, err = cli.GetInfo("v0.13.0")
			require.NoError(t, err)

			recovered := false
			for _, entry := range hook.AllEntries() {
				if entry.Message == "Recovered after retrying" {
					recovered = true
				}
			}

			assert.Equal(t, test.outage, recovered)
		})
	}
}

func TestTheServerAnswerIsBoundedBeforeItReachesTheLog(t *testing.T) {
	backend, hook := logtest.NewNullLogger()
	backend.SetLevel(logrus.DebugLevel)

	cli, err := NewClient("https://www.cloud.shellhub.io/", withImmediateRetries(), WithLogger(backend))
	require.NoError(t, err)

	client, ok := cli.(*client)
	require.True(t, ok)

	mock.ActivateNonDefault(client.http.GetClient())
	defer mock.DeactivateAndReset()

	page := strings.Repeat("<html>a proxy error page nobody wants a line of</html>", 100)

	accepted, _ := mock.NewJsonResponder(200, models.DeviceAuthResponse{Name: "83-18-77-25-78-0d"})
	mock.RegisterResponder("POST", "/api/devices/auth",
		mock.NewStringResponder(http.StatusBadGateway, page).Then(accepted))

	_, err = cli.AuthDevice(authRequest())
	require.NoError(t, err)

	require.NotEmpty(t, hook.AllEntries())

	reported, ok := hook.AllEntries()[0].Data[logrus.ErrorKey].(error)
	require.True(t, ok)

	assert.Less(t, len(reported.Error()), len(page))
	assert.Contains(t, reported.Error(), "502")
}

func TestAuthorizationStopsRetryingAfterItsDeadline(t *testing.T) {
	cli, err := NewClient("https://www.cloud.shellhub.io/", withImmediateRetries(), WithAuthorizationDeadline(time.Nanosecond))
	require.NoError(t, err)

	client, ok := cli.(*client)
	require.True(t, ok)

	mock.ActivateNonDefault(client.http.GetClient())
	defer mock.DeactivateAndReset()

	mock.RegisterResponder("POST", "/api/devices/auth",
		mock.NewStringResponder(http.StatusNotFound, `{"message":"namespace not found"}`))

	_, err = cli.AuthDevice(authRequest())
	require.ErrorIs(t, err, ErrNotFound)

	calls := 0
	for _, count := range mock.GetCallCountInfo() {
		calls += count
	}

	assert.Equal(t, 1, calls)
}

func TestAuthorizationKeepsRetryingWithinItsDeadline(t *testing.T) {
	cli, err := NewClient("https://www.cloud.shellhub.io/", withImmediateRetries(), WithAuthorizationDeadline(time.Hour))
	require.NoError(t, err)

	client, ok := cli.(*client)
	require.True(t, ok)

	mock.ActivateNonDefault(client.http.GetClient())
	defer mock.DeactivateAndReset()

	accepted, _ := mock.NewJsonResponder(200, models.DeviceAuthResponse{Name: "83-18-77-25-78-0d"})
	mock.RegisterResponder("POST", "/api/devices/auth",
		mock.NewStringResponder(http.StatusNotFound, `{"message":"namespace not found"}`).Then(accepted))

	response, err := cli.AuthDevice(authRequest())
	require.NoError(t, err)
	assert.NotNil(t, response)
}
