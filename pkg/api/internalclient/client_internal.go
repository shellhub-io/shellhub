package internalclient

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	apiHost    = "api"
	apiPort    = 8080
	apiScheme  = "http"
	billingURL = "billing-api"
)

type Client interface {
	commonAPI
	internalAPI
}

type internalAPI interface {
	LookupDevice()
	GetPublicKey(fingerprint, tenant string) (*models.PublicKey, error)
	CreatePrivateKey() (*models.PrivateKey, error)
	EvaluateKey(fingerprint string, dev *models.Device, username string) (bool, error)
	DevicesOffline(id string) error
	DevicesHeartbeat(id string) error
	FirewallEvaluate(lookup map[string]string) error
	SessionAsAuthenticated(uid string) []error
	FinishSession(uid string) []error
	KeepAliveSession(uid string) []error
	RecordSession(session *models.SessionRecorded, recordURL string)
	Lookup(lookup map[string]string) (string, []error)
	DeviceLookup(lookup map[string]string) (*models.Device, []error)
	BillingReport(tenant string, action string) (int, error)
	BillingEvaluate(tenantID string) (*models.BillingEvaluation, int, error)
}

func (c *client) LookupDevice() {
}

func (c *client) BillingReport(tenant string, action string) (int, error) {
	res, err := c.http.R().
		SetHeader("X-Tenant-ID", tenant).
		SetQueryParam("action", action).
		Post(fmt.Sprintf("%s://%s:%d/internal/billing/report", apiScheme, billingURL, apiPort))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return res.StatusCode(), nil
}

func (c *client) BillingEvaluate(tenantID string) (*models.BillingEvaluation, int, error) {
	var evaluation *models.BillingEvaluation
	resp, err := c.http.R().
		SetHeader("X-Tenant-ID", tenantID).
		SetResult(&evaluation).
		Post(fmt.Sprintf("%s://%s:%d/internal/billing/evaluate", apiScheme, billingURL, apiPort))
	if err != nil {
		return evaluation, resp.StatusCode(), err
	}

	return evaluation, resp.StatusCode(), nil
}

func (c *client) GetPublicKey(fingerprint, tenant string) (*models.PublicKey, error) {
	var pubKey *models.PublicKey
	resp, err := c.http.R().
		SetResult(&pubKey).
		Get(buildURL(c, fmt.Sprintf("/internal/sshkeys/public-keys/%s/%s", fingerprint, tenant)))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == 404 {
		return nil, ErrNotFound
	}

	return pubKey, nil
}

func (c *client) EvaluateKey(fingerprint string, dev *models.Device, username string) (bool, error) {
	var evaluate *bool

	resp, err := c.http.R().
		SetBody(dev).
		SetResult(&evaluate).
		Post(buildURL(c, fmt.Sprintf("/internal/sshkeys/public-keys/evaluate/%s/%s", fingerprint, username)))
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == 200 {
		return *evaluate, nil
	}

	return false, nil
}

func (c *client) CreatePrivateKey() (*models.PrivateKey, error) {
	var privKey *models.PrivateKey
	_, err := c.http.R().
		SetResult(&privKey).
		Post(buildURL(c, "/internal/sshkeys/private-keys"))
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func (c *client) DevicesOffline(id string) error {
	_, err := c.http.R().
		Post(buildURL(c, fmt.Sprintf("/internal/devices/%s/offline", id)))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) DevicesHeartbeat(id string) error {
	_, err := c.asynq.Enqueue(asynq.NewTask("api:heartbeat", []byte(id)), asynq.Queue("api"), asynq.Group("heartbeats"))

	return err
}

var (
	ErrFirewallConnection = errors.New("failed to make the request to evaluate the firewall")
	ErrFirewallBlock      = errors.New("a firewall rule prohibit this connection")
)

func (c *client) FirewallEvaluate(lookup map[string]string) error {
	local := resty.New()
	local.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok {
			return true
		}

		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	resp, err := local.
		SetRetryCount(10).
		R().
		SetQueryParams(lookup).
		Get("http://cloud-api:8080/internal/firewall/rules/evaluate")
	if err != nil {
		return ErrFirewallConnection
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrFirewallBlock
	}

	return nil
}

// SessionAsAuthenticated makes a HTTP request to ShellHub API server to mark the session as authenticated.
func (c *client) SessionAsAuthenticated(uid string) []error {
	var errors []error
	_, err := c.http.R().
		SetBody(&models.Status{
			Authenticated: true,
		}).
		Patch(buildURL(c, fmt.Sprintf("/internal/sessions/"+uid)))
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) FinishSession(uid string) []error {
	var errors []error
	_, err := c.http.R().
		Post(buildURL(c, fmt.Sprintf("/internal/sessions/%s/finish", uid)))
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) KeepAliveSession(uid string) []error {
	var errors []error
	_, err := c.http.R().
		Post(buildURL(c, fmt.Sprintf("/internal/sessions/%s/keepalive", uid)))
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) RecordSession(session *models.SessionRecorded, recordURL string) {
	_, _ = c.http.R().
		SetBody(session).
		Post(fmt.Sprintf("http://"+recordURL+"/internal/sessions/%s/record", session.UID))
}

func (c *client) Lookup(lookup map[string]string) (string, []error) {
	var device struct {
		UID string `json:"uid"`
	}

	resp, _ := c.http.R().
		SetQueryParams(lookup).
		SetResult(&device).
		Get(buildURL(c, "/internal/lookup"))

	if resp.StatusCode() != http.StatusOK {
		return "", []error{errors.New("lookup failed")}
	}

	return device.UID, nil
}

func (c *client) DeviceLookup(lookup map[string]string) (*models.Device, []error) {
	var device *models.Device

	resp, err := c.http.R().
		SetQueryParams(lookup).
		SetResult(&device).
		Get(buildURL(c, "/internal/lookup"))
	if err != nil {
		return nil, []error{err}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, []error{err}
	}

	return device, nil
}
