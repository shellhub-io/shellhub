package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
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
	NewReverseListener(ctx context.Context, token string, connPath string) (*revdial.Listener, error)
}

//go:generate mockery --name=Client --filename=client.go
type Client interface {
	commonAPI
	publicAPI
}

type client struct {
	scheme string
	host   string
	port   int
	http   *resty.Client
	logger *log.Logger
	// reverser is used to create a reverse listener to Agent from ShellHub's SSH server.
	reverser IReverser
}

var ErrParseAddress = fmt.Errorf("could not parse the address to the required format")

// NewClient creates a new ShellHub HTTP client.
//
// Server address must contain the scheme, the host and the port. For instance: `https://cloud.shellhub.io:443/`.
func NewClient(address string, opts ...Opt) (Client, error) {
	uri, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, errors.Join(ErrParseAddress, err)
	}

	const RetryAfterHeader string = "Retry-After"

	// DefaultMaxRetryWaitTime is the default value for wait time between retries.
	const DefaultMaxRetryWaitTime time.Duration = 1 * time.Hour
	// DefaultRetryAfterTime is the retry default time when the header [RetryAfterHeader] isn't defined on the response.
	const DefaultRetryAfterTime time.Duration = 5 * time.Second

	client := new(client)
	client.http = resty.New()
	client.http.SetRetryCount(math.MaxInt32)
	client.http.SetRedirectPolicy(SameDomainRedirectPolicy())
	client.http.SetBaseURL(uri.String())
	client.http.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok {
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
				return DefaultRetryAfterTime, nil
			}

			// NOTE: The `Retry-After` supports delay in seconds and and a date time, but currently we will support only
			// one of them.
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Retry-After
			retryAfterSeconds, err := strconv.Atoi(retryAfterHeader)
			if err != nil {
				return DefaultRetryAfterTime, err
			}

			log.WithFields(log.Fields{
				"status":      r.StatusCode(),
				"retry_after": retryAfterSeconds,
				"url":         r.Request.URL,
			}).Debug("retrying request after a defined time period")

			return time.Duration(retryAfterSeconds) * time.Second, nil
		default:
			return DefaultRetryAfterTime, nil
		}
	})
	client.http.SetRetryMaxWaitTime(DefaultMaxRetryWaitTime)

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
