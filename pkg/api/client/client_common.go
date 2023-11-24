package client

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"

	resty "github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

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

func (c *client) ListDevices() ([]models.Device, error) {
	devices := make([]models.Device, 0)

	response, err := c.http.R().
		SetResult(&devices).
		Get("/api/devices")
	if err != nil {
		return nil, err
	}

	if err := ErrorFromResponse(response); err != nil {
		return nil, err
	}

	return devices, nil
}

func (c *client) GetDevice(uid string) (*models.Device, error) {
	var device *models.Device

	response, err := c.http.R().
		SetResult(&device).
		Get(fmt.Sprintf("/api/devices/%s", uid))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	if err := ErrorFromResponse(response); err != nil {
		return nil, err
	}

	return device, nil
}
