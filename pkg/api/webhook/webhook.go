package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/go-resty/resty/v2"
	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrConnectionFailed = errors.New("connection failed")
	ErrForbidden        = errors.New("not allowed")
	ErrUnknown          = errors.New("unknown error")
)

type Webhook interface {
	Connect(m map[string]string) (*IncomingConnectionWebhookResponse, error)
}

type Options struct {
	WebhookURL    string `env:"WEBHOOK_URL"`
	WebhookPort   int    `env:"WEBHOOK_PORT"`
	WebhookScheme string `env:"WEBHOOK_SCHEME"`
}

func NewClient() Webhook {
	httpClient := resty.New()
	httpClient.SetRetryCount(3)
	opts, err := envs.ParseWithPrefix[Options]("")
	if err != nil {
		return nil
	}

	w := &webhookClient{
		host:   opts.WebhookURL,
		port:   opts.WebhookPort,
		scheme: opts.WebhookScheme,
		http:   httpClient,
	}

	if w.logger != nil {
		httpClient.SetLogger(&client.LeveledLogger{w.logger})
	}

	return w
}

type webhookClient struct {
	scheme string
	host   string
	port   int
	http   *resty.Client
	logger *logrus.Logger
}

func (w *webhookClient) Connect(m map[string]string) (*IncomingConnectionWebhookResponse, error) {
	payload := &IncomingConnectionWebhookRequest{
		Username:  m["username"],
		Hostname:  m["name"],
		Namespace: m["domain"],
		SourceIP:  m["ip_address"],
	}
	secret := "secret"
	uuid := uuid.Generate()
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(fmt.Sprintf("%v", payload))); err != nil {
		return nil, err
	}
	signature := mac.Sum(nil)

	var res *IncomingConnectionWebhookResponse
	resp, err := w.http.R().
		SetHeaders(map[string]string{
			WebhookIDHeader:        uuid,
			WebhookEventHeader:     WebhookIncomingConnectionEvent,
			WebhookSignatureHeader: hex.EncodeToString(signature),
		}).
		SetBody(payload).
		SetResult(&res).
		Post(buildURL(w, "/"))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	if resp.StatusCode() == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if resp.StatusCode() == http.StatusOK {
		return res, nil
	}

	return nil, ErrUnknown
}

func buildURL(w *webhookClient, uri string) string {
	u, _ := url.Parse(fmt.Sprintf("%s://%s:%d", w.scheme, w.host, w.port))
	u.Path = path.Join(u.Path, uri)

	return u.String()
}
