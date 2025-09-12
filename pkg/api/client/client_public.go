package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/hashicorp/yamux"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
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
		AddRetryCondition(func(r *resty.Response, _ error) bool {
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
					"data":        r.String(),
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

// NewReverseListener creates a new reverse listener connection to ShellHub's server. This listener receives the SSH
// requests coming from the ShellHub server. Only authenticated devices can obtain a listener connection.
func (c *client) NewReverseListenerV1(ctx context.Context, token string, path string) (net.Listener, error) {
	if token == "" {
		return nil, errors.New("token is empty")
	}

	if err := c.reverser.Auth(ctx, token, path); err != nil {
		return nil, err
	}

	return c.reverser.NewListener()
}

type ReverseListenerV2Config struct {
	// AcceptBacklog is used to limit how many streams may be
	// waiting an accept.
	AcceptBacklog int `json:"yamux_accept_backlog"`

	// EnableKeepalive is used to do a period keep alive
	// messages using a ping.
	EnableKeepAlive bool `json:"yamux_enable_keep_alive"`

	// KeepAliveInterval is how often to perform the keep alive
	KeepAliveInterval time.Duration `json:"yamux_keep_alive_interval"`

	// ConnectionWriteTimeout is meant to be a "safety valve" timeout after
	// we which will suspect a problem with the underlying connection and
	// close it. This is only applied to writes, where's there's generally
	// an expectation that things will move along quickly.
	ConnectionWriteTimeout time.Duration `json:"yamux_connection_write_timeout"`

	// MaxStreamWindowSize is used to control the maximum
	// window size that we allow for a stream.
	MaxStreamWindowSize uint32 `json:"yamux_max_stream_window_size"`

	// StreamOpenTimeout is the maximum amount of time that a stream will
	// be allowed to remain in pending state while waiting for an ack from the peer.
	// Once the timeout is reached the session will be gracefully closed.
	// A zero value disables the StreamOpenTimeout allowing unbounded
	// blocking on OpenStream calls.
	StreamOpenTimeout time.Duration `json:"yamux_stream_open_timeout"`

	// StreamCloseTimeout is the maximum time that a stream will allowed to
	// be in a half-closed state when `Close` is called before forcibly
	// closing the connection. Forcibly closed connections will empty the
	// receive buffer, drop any future packets received for that stream,
	// and send a RST to the remote side.
	StreamCloseTimeout time.Duration `json:"yamux_stream_close_timeout"`
}

var DefaultReverseListenerV2Config = ReverseListenerV2Config{
	AcceptBacklog:          256,
	EnableKeepAlive:        true,
	KeepAliveInterval:      35 * time.Second,
	ConnectionWriteTimeout: 15 * time.Second,
	MaxStreamWindowSize:    256 * 1024,
	StreamCloseTimeout:     5 * time.Minute,
	StreamOpenTimeout:      75 * time.Second,
}

// NewReverseV2ConfigFromMap creates a new Config from a map[string]any received from auth data from the server
// or returns the default config if the map is nil. If a key is missing, the default value is used.
func NewReverseV2ConfigFromMap(m map[string]any) *ReverseListenerV2Config {
	cfg := DefaultReverseListenerV2Config

	if v, ok := m["yamux_accept_backlog"].(int); ok {
		cfg.AcceptBacklog = v
	}

	if v, ok := m["yamux_enable_keep_alive"].(bool); ok {
		cfg.EnableKeepAlive = v
	}

	if v, ok := m["yamux_keep_alive_interval"].(time.Duration); ok {
		cfg.KeepAliveInterval = v
	}

	if v, ok := m["yamux_connection_write_timeout"].(time.Duration); ok {
		cfg.ConnectionWriteTimeout = v
	}

	if v, ok := m["yamux_max_stream_window_size"].(uint32); ok {
		cfg.MaxStreamWindowSize = v
	}

	if v, ok := m["yamux_stream_open_timeout"].(time.Duration); ok {
		cfg.StreamOpenTimeout = v
	}

	if v, ok := m["yamux_stream_close_timeout"].(time.Duration); ok {
		cfg.StreamCloseTimeout = v
	}

	return &cfg
}

func YamuxConfigFromReverseListenerV2(cfg *ReverseListenerV2Config) *yamux.Config {
	if cfg == nil {
		cfg = &DefaultReverseListenerV2Config
	}

	return &yamux.Config{
		AcceptBacklog:          cfg.AcceptBacklog,
		EnableKeepAlive:        cfg.EnableKeepAlive,
		KeepAliveInterval:      cfg.KeepAliveInterval,
		ConnectionWriteTimeout: cfg.ConnectionWriteTimeout,
		MaxStreamWindowSize:    cfg.MaxStreamWindowSize,
		StreamCloseTimeout:     cfg.StreamCloseTimeout,
		StreamOpenTimeout:      cfg.StreamOpenTimeout,
		// NOTE: LogOutput is required, and without it yamux will failed to create the session.
		LogOutput: os.Stderr,
	}
}

func (c *client) NewReverseListenerV2(ctx context.Context, token string, path string, cfg *ReverseListenerV2Config) (net.Listener, error) {
	if token == "" {
		return nil, errors.New("token is empty")
	}

	u, err := url.JoinPath(c.http.BaseURL, path)
	if err != nil {
		return nil, err
	}

	wsconn, _, err := DialContext(ctx, u, http.Header{
		"Authorization": []string{"Bearer " + token},
	})
	if err != nil {
		return nil, err
	}

	var listener *yamux.Session

	conn := wsconnadapter.New(wsconn)

	listener, err = yamux.Server(conn, YamuxConfigFromReverseListenerV2(cfg))
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"accept_backlog":           cfg.AcceptBacklog,
			"enable_keep_alive":        cfg.EnableKeepAlive,
			"keep_alive_interval":      cfg.KeepAliveInterval,
			"connection_write_timeout": cfg.ConnectionWriteTimeout,
			"max_stream_window_size":   cfg.MaxStreamWindowSize,
			"stream_close_timeout":     cfg.StreamCloseTimeout,
			"stream_open_timeout":      cfg.StreamOpenTimeout,
		}).Error("failed to create muxed session")

		// NOTE: If we fail to create the session, we should try again with the [DefaultConfig] as the client
		// could be using different settings.
		log.WithError(err).Warning("trying to create muxed session with default config")
		listener, err = yamux.Server(conn, YamuxConfigFromReverseListenerV2(&DefaultReverseListenerV2Config))
		if err != nil {
			log.WithError(err).Error("failed to create muxed session with default config")

			return nil, err
		}

		log.WithError(err).Warning("muxed session created with default config due to error with custom config")
	}

	return listener, err
}
