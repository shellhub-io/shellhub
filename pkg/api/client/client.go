package client

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/client/reverser"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

type commonAPI interface {
	ListDevices() ([]models.Device, error)
	GetDevice(uid string) (*models.Device, error)
}

type publicAPI interface {
	GetInfo(agentVersion string) (*models.Info, error)
	Endpoints() (*models.Endpoints, error)
	AuthDevice(req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	AuthPublicKey(req *models.PublicKeyAuthRequest, token string) (*models.PublicKeyAuthResponse, error)
	CreateDeviceLoginCode(token string) (*models.DeviceLoginCode, error)
	GetDeviceAuthStatus(token string) (*models.DeviceAuthStatus, error)
	CreateDevicePairing(req *models.DevicePairingRequest) (*models.DevicePairing, error)
	GetDevicePairingStatus(code string) (*models.DevicePairingStatus, error)
	NewReverseListenerV1(ctx context.Context, token string, path string) (net.Listener, error)
	NewReverseListenerV2(ctx context.Context, token string, path string, cfg *ReverseListenerV2Config) (net.Listener, error)
}

// Client is the agent's view of the ShellHub API: the routes an agent calls, plus the reverse
// listener it serves SSH on. Build one with NewClient.
type Client interface {
	commonAPI
	publicAPI
}

type client struct {
	scheme   string
	host     string
	port     int
	http     *resty.Client
	logger   *log.Logger
	reverser reverser.Reverser
}

// ErrParseAddress is returned by NewClient when the server address is not a URL carrying scheme,
// host and port.
var ErrParseAddress = errors.New("could not parse the address to the required format")

// NewClient creates a new ShellHub HTTP client.
//
// Server address must contain the scheme, the host and the port. For instance: `https://cloud.shellhub.io:443/`.
//
// The client retries indefinitely, backing off on the server's Retry-After when it sends one and on
// a random delay when it does not, so an agent left running against a server that is down reconnects
// on its own. Body-less requests go out with Content-Length: 0 rather than an empty chunked body,
// which proxies and request binders handle far more predictably.
func NewClient(address string, opts ...Opt) (Client, error) {
	uri, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, errors.Join(ErrParseAddress, err)
	}

	const RetryAfterHeader string = "Retry-After"

	const MaxRetryWaitTime time.Duration = 1 * time.Hour

	randomWaitTimeSecs := func() time.Duration {
		const MinRetryAfterSecs int = 5
		const MaxRetryAfterSecs int = 65

		t := time.Duration(rand.IntN(MaxRetryAfterSecs-MinRetryAfterSecs)+MinRetryAfterSecs) * time.Second //nolint:gosec

		log.WithFields(log.Fields{
			"retry_after": t,
		}).Warn("retrying request after a random time period")

		return t
	}

	client := new(client)
	client.http = resty.New()
	client.http.SetRetryCount(math.MaxInt32)
	client.http.SetRedirectPolicy(SameDomainRedirectPolicy())
	client.http.SetBaseURL(uri.String())
	client.http.SetContentLength(true)
	client.http.AddRetryCondition(func(r *resty.Response, err error) bool {
		var netErr net.Error
		if errors.As(err, &netErr) {
			log.WithFields(log.Fields{
				"url": r.Request.URL,
			}).WithError(err).Error("network error")

			return true
		}

		switch {
		case r.StatusCode() == http.StatusTooManyRequests:
			log.WithFields(log.Fields{
				"status_code": r.StatusCode(),
				"url":         r.Request.URL,
				"data":        r.String(),
			}).Warn("too many requests")

			return true
		case r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented:
			log.WithFields(log.Fields{
				"status_code": r.StatusCode(),
				"url":         r.Request.URL,
				"data":        r.String(),
			}).Warn("failed to achieve the server")

			return true
		}

		return false
	})
	client.http.SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
		switch r.StatusCode() {
		case http.StatusTooManyRequests, http.StatusServiceUnavailable:
			retryAfterHeader := r.Header().Get(RetryAfterHeader)
			if retryAfterHeader == "" {
				return randomWaitTimeSecs(), nil
			}

			retryAfterSeconds, err := strconv.Atoi(retryAfterHeader)
			if err != nil {
				return randomWaitTimeSecs(), err
			}

			log.WithFields(log.Fields{
				"status":      r.StatusCode(),
				"retry_after": retryAfterSeconds,
				"url":         r.Request.URL,
			}).Warn("retrying request after a defined time period")

			return time.Duration(retryAfterSeconds) * time.Second, nil
		default:
			return randomWaitTimeSecs(), nil
		}
	})
	client.http.SetRetryMaxWaitTime(MaxRetryWaitTime)

	if client.logger != nil {
		client.http.SetLogger(&LeveledLogger{client.logger})
	}

	client.reverser = NewReverser(client.http.BaseURL)

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}
