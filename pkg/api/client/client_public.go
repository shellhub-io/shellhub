package client

import (
	"context"
	"errors"

	resty "github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	log "github.com/sirupsen/logrus"
)

func (c *client) GetInfo(agentVersion string) (*models.Info, error) {
	var info *models.Info

	response, err := c.http.R().
		SetResult(&info).
		Get("/info?agent_version=" + agentVersion)
	if err != nil {
		return nil, err
	}

	if err := ErrorFromResponse(response); err != nil {
		return nil, err
	}

	return info, nil
}

func (c *client) AuthDevice(req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	var res *models.DeviceAuthResponse

	response, err := c.http.R().
		AddRetryCondition(func(r *resty.Response, err error) bool {
			identity := func(mac, hostname string) string {
				if mac != "" {
					return mac
				}

				return hostname
			}

			if r.IsError() {
				log.WithFields(log.Fields{
					"tenant_id":   req.TenantID,
					"identity":    identity(req.Identity.MAC, req.Hostname),
					"status_code": r.StatusCode(),
				}).Warn("failed to authenticate device")

				return true
			}

			return false
		}).
		SetBody(req).
		SetResult(&res).
		Post("/api/devices/auth")
	if err != nil {
		return nil, err
	}

	if err := ErrorFromResponse(response); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *client) Endpoints() (*models.Endpoints, error) {
	var endpoints *models.Endpoints

	response, err := c.http.R().
		SetResult(&endpoints).
		Get("/endpoints")
	if err != nil {
		return nil, err
	}

	if err := ErrorFromResponse(response); err != nil {
		return nil, err
	}

	return endpoints, nil
}

func (c *client) AuthPublicKey(req *models.PublicKeyAuthRequest, token string) (*models.PublicKeyAuthResponse, error) {
	var res *models.PublicKeyAuthResponse

	response, err := c.http.R().
		SetBody(req).
		SetResult(&res).
		SetAuthToken(token).
		Post("/api/auth/ssh")
	if err != nil {
		return nil, err
	}

	if err := ErrorFromResponse(response); err != nil {
		return nil, err
	}

	return res, nil
}

// NewReverseListener creates a new reverse listener connection for the Agent from ShellHub's SSH server.
//
// Every time the ShellHub's SSH server receives a new connection to the Agent, the server sends that connection
// through this listener to the device, but only if it is accepted by the server.
//
// To obtain this listener from the server, the Agent needs to authenticates it using the token provided, and getting a
// reverse authenticated connection, after that, it dials the server again for a new reverse connection on ShellHub's
// SSH tunnel list.
func (c *client) NewReverseListener(ctx context.Context, token string) (*revdial.Listener, error) {
	if token == "" {
		return nil, errors.New("token is empty")
	}

	if err := c.reverser.Auth(ctx, token); err != nil {
		return nil, err
	}

	return c.reverser.NewListener()
}
