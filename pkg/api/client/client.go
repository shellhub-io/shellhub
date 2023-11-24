package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"

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
	NewReverseListener(ctx context.Context, token string) (*revdial.Listener, error)
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

	client := new(client)
	client.http = resty.New()
	client.http.SetRetryCount(math.MaxInt32)
	client.http.SetRedirectPolicy(SameDomainRedirectPolicy())
	client.http.SetBaseURL(uri.String())
	client.http.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok {
			return true
		}

		if r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented {
			log.WithFields(log.Fields{
				"status_code": r.StatusCode(),
				"url":         r.Request.URL,
			}).Warn("failed to achieve the server")

			return true
		}

		return false
	})

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
